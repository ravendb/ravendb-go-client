package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

func getNextOperationIdCommandTest_canGetNextOperationId(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	command := ravendb.NewGetNextOperationIdCommand()
	err = store.GetRequestExecutor().ExecuteCommand(command)
	assert.NoError(t, err)
	assert.NotNil(t, command.Result)
}

func TestGetNextOperationIdCommand(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// follows execution order of java tests
	getNextOperationIdCommandTest_canGetNextOperationId(t)
}
