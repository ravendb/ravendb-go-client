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
	err = store.getRequestExecutor().executeCommand(command)
	assert.NoError(t, err)
	assert.NotNil(t, command.Result)
}

func TestGetNextOperationIdCommand(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer func() {
		r := recover()
		destroyDriver()
		if r != nil {
			panic(r)
		}
	}()
	// follows execution order of java tests
	getNextOperationIdCommandTest_canGetNextOperationId(t)
}
