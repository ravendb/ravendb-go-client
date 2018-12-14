package tests

import (
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func getTcpInfoTest_canGetTcpInfo(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	command := ravendb.NewGetTcpInfoCommand("test")
	err := store.GetRequestExecutor().ExecuteCommand(command)
	assert.NoError(t, err)
	result := command.Result
	assert.NotNil(t, result)
	assert.Nil(t, result.GetCertificate())
	// Note: in Java this tests for non-nil but Port is not sent
	// in Json, so don't quite understand that. Unless Java check
	// is bogus
	assert.Equal(t, 0, result.GetPort())
	assert.NotEmpty(t, result.GetUrl())
}

func TestGetTcpInfo(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	getTcpInfoTest_canGetTcpInfo(t, driver)
}
