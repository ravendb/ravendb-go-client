package ravendb

import (
	"fmt"
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
		if r := recover(); r != nil {
			fmt.Printf("Recovered in %s\n", t.Name())
		}
		destroyDriver()
	}()

	// follows execution order of java tests
	getNextOperationIdCommandTest_canGetNextOperationId(t)
}
