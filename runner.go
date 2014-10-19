package mrepo

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
)

//Seq run, in sequences the command on each project
// because of some commands optimisation, it is not the same as running them async, and then printing the output
// some commands DO not print the same output if they are connected to the stdout.
// besides, you lose the stdin ability.
func Seq(projects <-chan string, name string, args ...string) {
	var count int
	for prj := range projects {
		count++
		fmt.Printf("\033[00;32m%s\033[00m$ %s %s\n", prj, name, strings.Join(args, " "))
		cmd := exec.Command(name, args...)
		cmd.Dir = prj
		cmd.Stderr, cmd.Stdout, cmd.Stdin = os.Stderr, os.Stdout, os.Stdin
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error running '%s %s':\n    %s\n", name, strings.Join(args, " "), err.Error())
		}
	}
	fmt.Printf("Done (\033[00;32m%v\033[00m repositories)\n", count)
}

//List just count and print all directories.
func List(projects <-chan string) {
	var count int
	for prj := range projects {
		count++
		fmt.Printf("\033[00;32m%s\033[00m$ \n", prj)
	}
	fmt.Printf("Done (\033[00;32m%v\033[00m repositories)\n", count)
}

//Concurrent run, in sequences the command on each repository
// because of some commands optimisation, it is not the same as running them async, and then printing the output
// some commands DO not print the same output if they are connected to the stdout.
// besides, you lose the stdin ability.
func Concurrent(projects <-chan string, shouldPrint bool, outputF PostProcessor, name string, args ...string) {

	var slot string // a reserved space to print and delete messages
	if shouldPrint {
		slot = strings.Repeat(" ", 80)
		fmt.Printf("\033[00;32m%s\033[00m$ %s %s\n", "<for all>", name, strings.Join(args, " "))
	}

	outputer := make(chan execution)
	var waiter sync.WaitGroup
	for prj := range projects {
		waiter.Add(1)

		if shouldPrint {
			fmt.Print("\r    start ")
			if len(prj) > len(slot) {
				fmt.Printf("%s ...", prj[0:len(slot)])
			} else {
				fmt.Printf("%s ...%s", prj, slot[len(prj):])
			}
		}

		go func(prj string) {
			defer waiter.Done()
			cmd := exec.Command(name, args...)
			cmd.Dir = prj
			out, err := cmd.CombinedOutput()
			if err != nil {
				return
			}
			// keep
			//head := fmt.Sprintf("\033[00;32m%s\033[00m$ %s %s\n", prj, name, strings.Join(args, " "))
			//outputer <- head + string(out)
			outputer <- execution{Name: prj, Cmd: name, Args: args, Result: string(out)}
		}(prj)
	}
	if shouldPrint {
		fmt.Printf("\r    all started. waiting for tasks to complete...%s\n\n", slot)
	}

	go func() {
		waiter.Wait()
		close(outputer)
	}()
	outputF(outputer)

}
