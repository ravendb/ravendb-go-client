package tests

import (
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func getTcpInfoTestCanGetTcpInfo(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	command := ravendb.NewGetTcpInfoCommand("test", "")
	err := store.GetRequestExecutor("").ExecuteCommand(command, nil)
	assert.NoError(t, err)
	result := command.Result
	assert.NotNil(t, result)
	assert.Nil(t, result.Certificate)
	// Note: in Java this tests for non-nil but Port is not sent
	// in Json, so don't quite understand that. Unless Java check
	// is bogus
	assert.Equal(t, 0, result.Port)
	assert.NotEmpty(t, result.URL)
}

func TestGetTcpInfo(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	getTcpInfoTestCanGetTcpInfo(t, driver)
}
