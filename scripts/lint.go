package main

import (
	"fmt"
	"os/exec"
	"strings"
)

// This is a meta-linter to help code quality high.
// It runs several linters and shows the result.
// It applies filters over the results because otherwise
// there's too much noise.
// List of possible linters: https://github.com/alecthomas/gometalinter#supported-linters
// To run: go run scripts\lint.go

func runCommandAndPrintResult(cmd *exec.Cmd, installCmd func() error) error {
	d, err := cmd.CombinedOutput()
	if err != nil && installCmd != nil && strings.Contains(err.Error(), "executable file not found") {
		err2 := installCmd()
		if err2 != nil {
			tool := strings.Join(cmd.Args, " ")
			fmt.Printf("Running install command for tool '%s' failed with %s\n", tool, err2)
			return err
		}
		// re-run command after installing it
		cmd = exec.Command(cmd.Args[0], cmd.Args[1:]...)
		d, err = cmd.CombinedOutput()
	}
	if len(d) > 0 {
		tool := strings.Join(cmd.Args, " ")
		fmt.Printf("%s:\n%s\n", tool, d)
	}
	return err
}

func runGoVet() {
	cmd := exec.Command("go", "vet")
	err := runCommandAndPrintResult(cmd, nil)

	ignoreError := func(err error) bool {
		if err == nil {
			return true
		}
		s := err.Error()
		// this just means go vet found errors
		return strings.Contains(s, "exit status 2")
	}
	if !ignoreError(err) {
		tool := strings.Join(cmd.Args, " ")
		fmt.Printf("Running %s failed with %s\n", tool, err)
	}
}

func runGoVetShadow() {
	cmd := exec.Command("go", "tool", "vet", "-shadow", ".")
	err := runCommandAndPrintResult(cmd, nil)
	ignoreError := func(err error) bool {
		return err == nil
	}
	if !ignoreError(err) {
		tool := strings.Join(cmd.Args, " ")
		fmt.Printf("Running %s failed with %s\n", tool, err)
	}
}

func runDeadcode() {
	installDeadcode := func() error {
		cmd := exec.Command("go", "get", "-u", "github.com/tsenart/deadcode")
		return cmd.Run()
	}

	cmd := exec.Command("deadcode")
	err := runCommandAndPrintResult(cmd, installDeadcode)
	ignoreError := func(err error) bool {
		return err == nil
	}
	if !ignoreError(err) {
		tool := strings.Join(cmd.Args, " ")
		fmt.Printf("Running %s failed with %s\n", tool, err)
	}
}

func main() {
	//runGoVet()
	//runGoVetShadow()
	// TODO: gotype .
	// TODO: gotype -x .
	runDeadcode()
	// TODO: rest of linters
}
