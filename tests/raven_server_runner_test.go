package tests

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

func RavenServerRunner_run(locator *RavenServerLocator, logsDir string) (*Process, error) {
	processStartInfo, err := getProcessStartInfo(locator, logsDir)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(processStartInfo.command, processStartInfo.arguments...)
	stdoutReader, err := cmd.StdoutPipe()

	if false && ravenServerVerbose {
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

func getProcessStartInfo(locator *RavenServerLocator, logsDir string) (*ProcessStartInfo, error) {
	path := locator.serverPath
	if !fileExists(path) {
		return nil, fmt.Errorf("Server file was not found: %s", path)
	}
	commandArguments := []string{
		"--RunInMemory=true",
		"--License.Eula.Accepted=true",
		"--Setup.Mode=None",
		"--Testing.ParentProcessId=" + getProcessId(),
		// "--non-interactive",
	}
	if logsDir != "" {
		{
			arg := "--Logs.Path=" + logsDir
			commandArguments = append(commandArguments, arg)
		}
		{
			// modes: None, Operations, Information
			arg := "--Logs.Mode=Information"
			commandArguments = append(commandArguments, arg)
		}
	}
	if ravenServerVerbose {
		if logsDir == "" {
			arg := "--Logs.Mode=Information"
			commandArguments = append(commandArguments, arg)
		}
		commandArguments = append(commandArguments, "--log-to-console")
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
