package mrepo

import (
	"fmt"
	"os/exec"
	"strings"
)

//this file contains function do deal with git commands.

const (
	defaultTrimCut = "\n \t"
)

//GitBranch extract the current branch name (HEAD)
func GitBranch(prj string) (branch string, err error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = prj
	out, err := cmd.CombinedOutput()
	if err != nil {
		return
	}
	result := strings.Trim(string(out), defaultTrimCut)
	return result, nil
}

//GitPull automate the pull with ff only option
func GitPull(prj string) (result string, err error) {
	cmd := exec.Command("git", "pull", "--ff-only")
	cmd.Dir = prj
	out, err := cmd.CombinedOutput()
	result = strings.Trim(string(out), defaultTrimCut)
	if err != nil {
		return result, fmt.Errorf("failed to: %s$ git pull --ff-only : %s", prj, err.Error())
	}
	return result, nil
}

//GitClone clone a repo
func GitClone(wd, rel, remote, branch string) (result string, err error) {
	cmd := exec.Command("git", "clone", remote, "-b", branch, rel)
	cmd.Dir = wd
	out, err := cmd.CombinedOutput()
	result = strings.Trim(string(out), "\n \t")
	if err != nil {
		return result, fmt.Errorf("failed to git clone %s -b %s %s: %s", remote, branch, rel, err.Error())
	}
	return result, nil
}

//GitRemoteOrigin returns the current remote.origin.url
// if there is no "origin" remote, then an error is returned.
func GitRemoteOrigin(prj string) (origin string, err error) {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = prj
	out, err := cmd.CombinedOutput()
	if err != nil {
		return
	}
	result := strings.Trim(string(out), defaultTrimCut)
	return result, nil
}

//GitRevParseHead read the current commit sha1
func GitRevParseHead(prj string) (result string, err error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = prj
	out, err := cmd.CombinedOutput()
	result = strings.Trim(string(out), defaultTrimCut)
	if err != nil {
		return result, fmt.Errorf("failed to: %s$ git rev-parse HEAD : %s", prj, err.Error())
	}
	return result, nil
}
