package main

import (
	"fmt"
	"os"
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

func ignoreExitStatusError(err error) bool {
	if err == nil {
		return true
	}
	s := err.Error()
	// many lint tools return exit code 1 or 2 to indicate they found errors
	if strings.Contains(s, "exit status 1") {
		return true
	}
	return strings.Contains(s, "exit status 2")
}

func ignoreError(err error) bool {
	return err == nil
}

func goVet() {
	cmd := exec.Command("go", "vet")
	err := runCommandAndPrintResult(cmd, nil)

	if !ignoreExitStatusError(err) {
		tool := strings.Join(cmd.Args, " ")
		fmt.Printf("Running %s failed with %s\n", tool, err)
	}
}

func goVetShadow() {
	cmd := exec.Command("go", "tool", "vet", "-shadow", ".")
	err := runCommandAndPrintResult(cmd, nil)
	if !ignoreError(err) {
		tool := strings.Join(cmd.Args, " ")
		fmt.Printf("Running %s failed with %s\n", tool, err)
	}
}

func deadcode() {
	install := func() error {
		cmd := exec.Command("go", "get", "-u", "github.com/tsenart/deadcode")
		return cmd.Run()
	}

	cmd := exec.Command("deadcode")
	err := runCommandAndPrintResult(cmd, install)
	if !ignoreExitStatusError(err) {
		tool := strings.Join(cmd.Args, " ")
		fmt.Printf("Running %s failed with %s\n", tool, err)
	}
}

func varcheck() {
	install := func() error {
		cmd := exec.Command("go", "get", "-u", "gitlab.com/opennota/check/cmd/varcheck")
		return cmd.Run()
	}

	cmd := exec.Command("varcheck")
	err := runCommandAndPrintResult(cmd, install)
	if !ignoreExitStatusError(err) {
		tool := strings.Join(cmd.Args, " ")
		fmt.Printf("Running %s failed with %s\n", tool, err)
	}
}

func structcheck() {
	install := func() error {
		cmd := exec.Command("go", "get", "-u", "gitlab.com/opennota/check/cmd/structcheck")
		return cmd.Run()
	}

	cmd := exec.Command("structcheck")
	err := runCommandAndPrintResult(cmd, install)
	if !ignoreExitStatusError(err) {
		tool := strings.Join(cmd.Args, " ")
		fmt.Printf("Running %s failed with %s\n", tool, err)
	}
}

func aligncheck() {
	install := func() error {
		cmd := exec.Command("go", "get", "-u", "gitlab.com/opennota/check/cmd/aligncheck")
		return cmd.Run()
	}

	cmd := exec.Command("aligncheck")
	err := runCommandAndPrintResult(cmd, install)
	if !ignoreExitStatusError(err) {
		tool := strings.Join(cmd.Args, " ")
		fmt.Printf("Running %s failed with %s\n", tool, err)
	}
}

func megacheck() {
	install := func() error {
		cmd := exec.Command("go", "get", "-u", "honnef.co/go/tools/cmd/megacheck")
		return cmd.Run()
	}

	cmd := exec.Command("megacheck")
	err := runCommandAndPrintResult(cmd, install)
	if !ignoreExitStatusError(err) {
		tool := strings.Join(cmd.Args, " ")
		fmt.Printf("Running %s failed with %s\n", tool, err)
	}
}

func maligned() {
	install := func() error {
		cmd := exec.Command("go", "get", "-u", "github.com/mdempsky/maligned")
		return cmd.Run()
	}

	cmd := exec.Command("maligned")
	err := runCommandAndPrintResult(cmd, install)
	if !ignoreExitStatusError(err) {
		tool := strings.Join(cmd.Args, " ")
		fmt.Printf("Running %s failed with %s\n", tool, err)
	}
}

func errcheck() {
	install := func() error {
		cmd := exec.Command("go", "get", "-u", "github.com/kisielk/errcheck")
		return cmd.Run()
	}

	cmd := exec.Command("errcheck")
	err := runCommandAndPrintResult(cmd, install)
	if !ignoreExitStatusError(err) {
		tool := strings.Join(cmd.Args, " ")
		fmt.Printf("Running %s failed with %s\n", tool, err)
	}
}

func dupl() {
	install := func() error {
		cmd := exec.Command("go", "get", "-u", "github.com/mibk/dupl")
		return cmd.Run()
	}

	cmd := exec.Command("dupl")
	err := runCommandAndPrintResult(cmd, install)
	if !ignoreExitStatusError(err) {
		tool := strings.Join(cmd.Args, " ")
		fmt.Printf("Running %s failed with %s\n", tool, err)
	}
}

func ineffassign() {
	install := func() error {
		cmd := exec.Command("go", "get", "-u", "github.com/gordonklaus/ineffassign")
		return cmd.Run()
	}

	cmd := exec.Command("ineffassign", ".")
	err := runCommandAndPrintResult(cmd, install)
	if !ignoreExitStatusError(err) {
		tool := strings.Join(cmd.Args, " ")
		fmt.Printf("Running %s failed with %s\n", tool, err)
	}
}

func unconvert() {
	install := func() error {
		cmd := exec.Command("go", "get", "-u", "github.com/mdempsky/unconvert")
		return cmd.Run()
	}

	cmd := exec.Command("unconvert")
	err := runCommandAndPrintResult(cmd, install)
	if !ignoreExitStatusError(err) {
		tool := strings.Join(cmd.Args, " ")
		fmt.Printf("Running %s failed with %s\n", tool, err)
	}
}

func goconst() {
	install := func() error {
		cmd := exec.Command("go", "get", "-u", "github.com/jgautheron/goconst/cmd/goconst")
		return cmd.Run()
	}

	cmd := exec.Command("goconst", "./...")
	err := runCommandAndPrintResult(cmd, install)
	if !ignoreExitStatusError(err) {
		tool := strings.Join(cmd.Args, " ")
		fmt.Printf("Running %s failed with %s\n", tool, err)
	}
}

func misspell() {
	install := func() error {
		cmd := exec.Command("go", "get", "-u", "github.com/client9/misspell/cmd/misspell")
		return cmd.Run()
	}

	cmd := exec.Command("misspell", ".")
	err := runCommandAndPrintResult(cmd, install)
	if !ignoreExitStatusError(err) {
		tool := strings.Join(cmd.Args, " ")
		fmt.Printf("Running %s failed with %s\n", tool, err)
	}
}

func nakedret() {
	install := func() error {
		cmd := exec.Command("go", "get", "-u", "github.com/alexkohler/nakedret")
		return cmd.Run()
	}

	cmd := exec.Command("nakedret", ".")
	err := runCommandAndPrintResult(cmd, install)
	if !ignoreExitStatusError(err) {
		tool := strings.Join(cmd.Args, " ")
		fmt.Printf("Running %s failed with %s\n", tool, err)
	}
}

func runToolByName(tool string) {
	switch tool {
	case "vet", "govet":
		goVet()
	case "vetshadow", "shadow":
		goVetShadow()
	case "dead", "deadcode":
		deadcode()
	case "var", "varcheck":
		varcheck()
	case "align", "aligncheck":
		aligncheck()
	case "struct", "structcheck":
		structcheck()
	case "mega", "megacheck":
		megacheck()
	case "err", "errcheck":
		errcheck()
	case "dupl", "duplicate":
		dupl()
	case "assign", "ineffasign":
		ineffassign()
	case "unconvert":
		unconvert()
	case "const", "goconst":
		goconst()
	case "spell", "misspell":
		misspell()
	case "ret", "nakedret":
		nakedret()
	default:
		fmt.Printf("Unknown tool '%s'\n", tool)
		os.Exit(1)
	}
}

func main() {
	// either run tools specified in command line or all of them
	if len(os.Args) > 1 {
		for _, tool := range os.Args[1:] {
			runToolByName(tool)
		}
		return
	}

	goVet()
	goVetShadow()
	deadcode()
	varcheck()
	structcheck()
	aligncheck()
	megacheck()
	maligned()
	errcheck()
	//dupl() // too many results and mostly false positive
	ineffassign()
	unconvert()
	goconst()
	misspell()
	nakedret()

	// TODO: https://github.com/securego/gosec doesn't yet support
	// code outside GOPATH

}
