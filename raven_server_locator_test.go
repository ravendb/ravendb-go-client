package ravendb

import (
	"fmt"
	"os"
)

const (
	envServerPath = "RAVENDB_JAVA_TEST_SERVER_PATH"
)

type RavenServerLocator struct {
	serverPath       string
	commandArguments []string
	command          string
}

func NewRavenServerLocator() (*RavenServerLocator, error) {
	path := os.Getenv(envServerPath)
	if path == "" {
		return nil, fmt.Errorf("Unable to find RavenDB server path. Please make sure %s environment variable is set and is valid (current value = %v)", envServerPath, path)
	}
	return &RavenServerLocator{
		serverPath: path,
		command:    path,
	}, nil
}
