package ravendb

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
)

type Process struct {
	cmd          *exec.Cmd
	stdoutReader io.ReadCloser
}

func RavenServerRunner_run(locator *RavenServerLocator) (*Process, error) {
	processStartInfo, err := getProcessStartInfo(locator)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(processStartInfo.command, processStartInfo.arguments...)
	stdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		fmt.Printf("exec.Command(%s, %v) failed with %s\n", processStartInfo.command, processStartInfo.arguments, err)
		return nil, err
	}
	return &Process{
		cmd:          cmd,
		stdoutReader: stdoutReader,
	}, nil
}

func getProcessStartInfo(locator *RavenServerLocator) (*ProcessStartInfo, error) {
	path := locator.serverPath
	if !fileExists(path) {
		return nil, fmt.Errorf("Serer file was not found: %s", path)
	}
	commandArguments := []string{
		"--RunInMemory=true",
		"--License.Eula.Accepted=true",
		"--Setup.Mode=None",
		"--Testing.ParentProcessId=" + getProcessId(),
	}
	commandArguments = append(locator.commandArguments, commandArguments...)
	res := &ProcessStartInfo{
		command:   locator.command,
		arguments: commandArguments,
	}
	return res, nil
}

func getProcessId() string {
	pid := os.Getpid()
	return strconv.Itoa(pid)
}

type ProcessStartInfo struct {
	command   string
	arguments []string
}
