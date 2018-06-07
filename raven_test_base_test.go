package ravendb

import (
	"fmt"
	"os"
)

func NewTestServiceLocator() (*RavenServerLocator, error) {
	locator, err := NewRavenServerLocator()
	if err != nil {
		return nil, err
	}
	locator.commandArguments = []string{
		"--ServerUrl=http://127.0.0.1:0",
	}
	return locator, nil
}

const (
	envCertificatePath = "RAVENDB_JAVA_TEST_CERTIFICATE_PATH"
	envHTTPSServerURL  = "RAVENDB_JAVA_TEST_HTTPS_SERVER_URL"
)

func NewSecuredServiceLocator() (*RavenServerLocator, error) {
	locator, err := NewRavenServerLocator()
	if err != nil {
		return nil, err
	}
	httpsServerURL := os.Getenv(envHTTPSServerURL)
	if httpsServerURL == "" {
		return nil, fmt.Errorf("Unable to find RavenDB https server url. Please make sure %s environment variable is set and is valid (current value = %v)", envHTTPSServerURL, envHTTPSServerURL)
	}

	certificatePath := os.Getenv(envCertificatePath)
	if certificatePath == "" {
		return nil, fmt.Errorf("Unable to find RavenDB server certificate path. Please make sure %s environment variable is set and is valid (current value = %v)", envCertificatePath, envHTTPSServerURL)
	}
	locator.commandArguments = []string{
		"--Security.Certificate.Path=" + certificatePath,
		"--ServerUrl=" + httpsServerURL,
	}
	return locator, nil
}

// TODO:     protected ConfigureRevisionsOperation.ConfigureRevisionsOperationResult setupRevisions(IDocumentStore store, boolean purgeOnDelete, long minimumRevisionsToKeep) {