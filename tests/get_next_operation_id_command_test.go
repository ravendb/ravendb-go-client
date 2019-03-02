package tests

import (
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func getNextOperationIdCommandTestCanGetNextOperationId(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	command := ravendb.NewGetNextOperationIDCommand()
	err = store.GetRequestExecutor("").ExecuteCommand(command, nil)
	assert.NoError(t, err)
	assert.NotNil(t, command.Result)
}

func TestGetNextOperationIDCommand(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// follows execution order of java tests
	getNextOperationIdCommandTestCanGetNextOperationId(t, driver)
}
