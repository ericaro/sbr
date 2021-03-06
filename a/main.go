package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ericaro/sbr/cmd"
	"github.com/ericaro/sbr/sbr"
)

const (
	Usage = `USAGE %[1]s [-options] <command> <args...>
			
DESCRIPTION:

  Run '<command> <args...>' is every repository found in the current directory hierarchy.

OPTIONS:
	
`
	Example = `
EXAMPLE

%[1]s git status -s
`
)

// the main that run a command on all sub commands
var (
	async = flag.Bool("a", false, "Controls the execution mode.\n           '-a' or '-a=true' run commands asynchronously.\n           '-a=false' of by default run commands sequentially.")
	list  = flag.Bool("l", false, "Dry mode just list the repositories.")

	// output selection
	cat    = flag.Bool("cat", false, "concatenate outputs, and print it")
	sum    = flag.Bool("sum", false, "parse each output as a number and print out the total")
	count  = flag.Bool("count", false, "count different outputs, and prints the resulting histogram")
	digest = flag.Bool("digest", false, "compute the sha1 digest of all outputs")

	// missing an outputer that takes care of "error codes"

	help = flag.Bool("h", false, "Print this help.")
)

func usage() {
	fmt.Printf(Usage, os.Args[0])
	flag.PrintDefaults()
	fmt.Printf(Example, os.Args[0])
}

func main() {
	flag.Parse()
	if (flag.NArg() == 0 && !*list) || *help {
		usage()
		os.Exit(-1)
	}

	// use wd by default
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error, cannot determine the current directory. %s\n", err.Error())
	}
	//build the workspace, that is used to trigger all commands
	workspace := sbr.NewWorkspace(wd)

	// parses the remaining args in order to pass them to the underlying process
	args := make([]string, 0)
	if flag.NArg() > 1 {
		args = flag.Args()[1:]
	}
	name := flag.Arg(0)

	if *list {
		//for now there is only one way to print dependencies
		//List just count and print all directories.
		var count int
		for _, prj := range workspace.ScanRel() {
			count++
			rel, err := filepath.Rel(wd, prj)
			if err != nil {
				rel = prj // uses the absolute path in this case
			}
			fmt.Printf("\033[00;32m%s\033[00m$ \n", rel)
		}
		fmt.Printf("Done (\033[00;32m%v\033[00m repositories)\n", count)

	} else {
		// select the output mode

		//again, passing the stdin, and stdout to the subprocess prevent: async, and ability to collect the outputs
		// for special outputers we need to collect outputs, so the 'special' var.
		// special => concurrent mode (because we need to collect outputs)
		// Therefore, selecting the output mode imply selecting "special"= true|false
		// and the ExecutionProcessor function
		var special bool = true
		var xp cmd.ExecutionProcessor
		switch {
		case *cat:
			xp = cmd.ExecutionCat
		case *sum:
			xp = cmd.ExecutionSum
		case *count:
			xp = cmd.ExecutionCount
		case *digest:
			xp = cmd.ExecutionDigest
		default:
			xp = cmd.ExecutionPrinter
			special = false
		}
		if special || *async { // this implies concurrent
			// based on the async option, exec asynchronously or sequentially.
			// we cannot just make "seq" a special case of concurrent, since when running sequentially we provide
			// direct access to the std streams. commands can use stdin, and use term escape codes.
			// When in async mode, we just can't do that.
			executions := cmd.ExecConcurrently(workspace, name, args...)
			xp(executions)

		} else {
			ExecSequentially(workspace, name, args...)
		}
	}

}

//ExecSequentially, for each `subrepository` in the working dir, execute the  command `command` with arguments `args`.
// It passes the stdin, stdout, and stderr to the subprocess. and wait for the result, before moving to the next one.
func ExecSequentially(x *sbr.Workspace, command string, args ...string) {
	var count int

	for _, sub := range x.ScanRel() {
		count++

		rel, err := filepath.Rel(x.Wd(), sub)
		if err != nil {
			rel = sub // use absolute path
		}
		//log.Printf("%v, %v", sub, rel)
		fmt.Printf("\033[00;32m%s\033[00m$ %s %s\n", rel, command, strings.Join(args, " "))
		cmd := exec.Command(command, args...)
		cmd.Dir = sub
		cmd.Stderr, cmd.Stdout, cmd.Stdin = os.Stderr, os.Stdout, os.Stdin
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error running '%s %s':\n    %s\n", command, strings.Join(args, " "), err.Error())
		}
	}
	fmt.Printf("Done (\033[00;32m%v\033[00m repositories)\n", count)
}
