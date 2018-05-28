package ravendb

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

func RavenServerRunner_run(locator *RavenServerLocator) (*exec.Cmd, error) {
	processStartInfo, err := getProcessStartInfo(locator)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(processStartInfo.command, processStartInfo.arguments...)
	err = cmd.Start()
	if err != nil {
		return nil, err
	}
	return cmd, nil
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
