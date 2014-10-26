package mrepo

import (
	"os"
	"path/filepath"
)

//scanner object to scan for a directory looking for git repositories.
type scanner struct {
	prjc chan string
	wd   string
}

//newScan creates a scanner
func newScan(workingDir string) *scanner {
	return &scanner{
		wd:   workingDir,
		prjc: make(chan string),
	}

}

//Find starts the directory scanning, and publish repository found.
func (s scanner) Find() (err error) {
	defer close(s.prjc)

	//I would like to do:
	//err = filepath.Walk(s.wd, s.walkFn)
	// but for backward compatibility (with 1.0.3) I can't call a method
	f := func(path string, f os.FileInfo, err error) error { return s.walkFn(path, f, err) }

	return filepath.Walk(s.wd, f)
}

//Repositories exposes the chan of repository.
// The chan is closed at the end.
func (s scanner) Repositories() <-chan string {
	return s.prjc
}

//WaldirFn compatible
func (s scanner) walkFn(path string, f os.FileInfo, err error) error {
	// if this path is a "prj", add it.
	// it would be something like if it's .git => it's parent is

	if f.IsDir() {
		if f.Name() == ".git" {
			// it's a repository file
			s.prjc <- filepath.Dir(path)
			//always skip the repository file
			return filepath.SkipDir
		}
	}
	return nil
}