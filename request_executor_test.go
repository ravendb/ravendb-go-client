package ravendb

import (
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

func RequestExecutorTest_failuresDoesNotBlockConnectionPool(t *testing.T) {
	conventions := NewDocumentConventions()
	store := getDocumentStoreMust(t)
	{
		executor := RequestExecutor_create(store.getUrls(), "no_such_db", nil, conventions)
		errorsCount := 0

		for i := 0; i < 40; i++ {
			command := NewGetNextOperationIdCommand()
			err := executor.executeCommand(command)
			if err != nil {
				errorsCount++
			}
		}
		assert.Equal(t, 40, errorsCount)

		databaseNamesOperation := NewGetDatabaseNamesOperation(0, 20)
		command := databaseNamesOperation.getCommand(conventions)
		err := executor.executeCommand(command)
		_ = err.(*DatabaseDoesNotExistException)
	}
}

func RequestExecutorTest_canIssueManyRequests(t *testing.T) {
	conventions := NewDocumentConventions()
	store := getDocumentStoreMust(t)
	{
		executor := RequestExecutor_create(store.getUrls(), store.getDatabase(), nil, conventions)
		for i := 0; i < 50; i++ {
			databaseNamesOperation := NewGetDatabaseNamesOperation(0, 20)
			command := databaseNamesOperation.getCommand(conventions)
			err := executor.executeCommand(command)
			assert.NoError(t, err)
		}
	}
}

func RequestExecutorTest_canFetchDatabasesNames(t *testing.T) {
	conventions := NewDocumentConventions()
	store := getDocumentStoreMust(t)
	{
		executor := RequestExecutor_create(store.getUrls(), store.getDatabase(), nil, conventions)

		databaseNamesOperation := NewGetDatabaseNamesOperation(0, 20)
		command := databaseNamesOperation.getCommand(conventions)
		err := executor.executeCommand(command)
		assert.NoError(t, err)

		dbNames := command.getResult().([]string)
		assert.True(t, stringArrayContains(dbNames, store.getDatabase()))
	}
}

func RequestExecutorTest_throwsWhenUpdatingTopologyOfNotExistingDb(t *testing.T) {
	conventions := NewDocumentConventions()
	store := getDocumentStoreMust(t)
	{
		executor := RequestExecutor_create(store.getUrls(), "no_such_db", nil, conventions)
		serverNode := NewServerNode()
		serverNode.setUrl(store.getUrls()[0])
		serverNode.setDatabase("no_such")
		future := executor.updateTopologyAsync(serverNode, 5000)
		_, err := future.get()
		_ = err.(*DatabaseDoesNotExistException)
	}
}

func RequestExecutorTest_throwsWhenDatabaseDoesNotExist(t *testing.T) {
	conventions := NewDocumentConventions()
	store := getDocumentStoreMust(t)
	{
		executor := RequestExecutor_create(store.getUrls(), "no_such_db", nil, conventions)
		command := NewGetNextOperationIdCommand()
		err := executor.executeCommand(command)
		_ = err.(*DatabaseDoesNotExistException)
	}
}

func RequestExecutorTest_canCreateSingleNodeRequestExecutor(t *testing.T) {
	//store := getDocumentStoreMust(t)
}

func RequestExecutorTest_canChooseOnlineNode(t *testing.T) {
	//store := getDocumentStoreMust(t)
}

func RequestExecutorTest_failsWhenServerIsOffline(t *testing.T) {
	//store := getDocumentStoreMust(t)
}

func TestRequestExecutor(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	if useProxy() {
		proxy.ChangeLogFile("trace_request_executor_go.txt")
	}

	// matches order of Java tests
	RequestExecutorTest_canFetchDatabasesNames(t)
	RequestExecutorTest_canIssueManyRequests(t)
	RequestExecutorTest_throwsWhenDatabaseDoesNotExist(t)

	RequestExecutorTest_failuresDoesNotBlockConnectionPool(t)

	RequestExecutorTest_canCreateSingleNodeRequestExecutor(t)
	RequestExecutorTest_failsWhenServerIsOffline(t)
	RequestExecutorTest_throwsWhenUpdatingTopologyOfNotExistingDb(t)
	RequestExecutorTest_canChooseOnlineNode(t)
}
