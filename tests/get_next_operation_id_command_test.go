package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

func getNextOperationIdCommandTest_canGetNextOperationId(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	command := ravendb.NewGetNextOperationIDCommand()
	err = store.GetRequestExecutor().ExecuteCommand(command)
	assert.NoError(t, err)
	assert.NotNil(t, command.Result)
}

func TestGetNextOperationIdCommand(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// follows execution order of java tests
	getNextOperationIdCommandTest_canGetNextOperationId(t, driver)
}
