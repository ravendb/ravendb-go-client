package ravendb

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
)

var (
	// RavenServerVerbose will cause raven server output to go to os.Stdout
	// if set to true. For debugging.
	RavenServerVerbose bool
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

	if false && RavenServerVerbose {
		cmd.Stderr = os.Stderr
		// cmd.StdoutPipe() sets cmd.Stdout to a pipe writer
		// we multi-plex it into os.Stdout
		// TODO: this doesn't seem to work. It makes reading from stdoutReader
		// immediately fail. Maybe it's becuse writer returned by
		// os.Pipe() (cmd.Stdout) blocks and MultiWriter() doesn't
		cmd.Stdout = io.MultiWriter(cmd.Stdout, os.Stdout)
	}
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
		"--non-interactive",
	}
	if RavenServerVerbose {
		commandArguments = append(commandArguments, "--log-to-console", "--Logs.Mode=Information", "--Logs.Path=./logs")
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