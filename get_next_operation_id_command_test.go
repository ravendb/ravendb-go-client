package ravendb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func getNextOperationIdCommandTest_canGetNextOperationId(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	command := NewGetNextOperationIdCommand()
	err = store.GetRequestExecutor().executeCommand(command)
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
