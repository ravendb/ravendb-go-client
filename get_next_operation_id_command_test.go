package ravendb

import (
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

func getNextOperationIdCommandTest_canGetNextOperationId(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)

	command := NewGetNextOperationIdCommand()
	err = store.getRequestExecutor().executeCommand(command)
	assert.NoError(t, err)
	assert.NotNil(t, command.Result)
}

func TestGetNextOperationIdCommand(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	if useProxy() {
		proxy.ChangeLogFile("trace_get_next_operation_id_go.txt")
	}

	createTestDriver()
	defer deleteTestDriver()

	// follows execution order of java tests
	getNextOperationIdCommandTest_canGetNextOperationId(t)
}
