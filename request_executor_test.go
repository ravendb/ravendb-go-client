package ravendb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	dbgRequestExecutorTests = false
)

func requestExecutorTest_failuresDoesNotBlockConnectionPool(t *testing.T) {
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_failuresDoesNotBlockConnectionPool start\n")
	}
	conventions := NewDocumentConventions()
	store := getDocumentStoreMust(t)
	defer store.Close()

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
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_failuresDoesNotBlockConnectionPool end\n")
	}
}

func requestExecutorTest_canIssueManyRequests(t *testing.T) {
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_canIssueManyRequests start\n")
	}
	conventions := NewDocumentConventions()
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		executor := RequestExecutor_create(store.getUrls(), store.getDatabase(), nil, conventions)
		for i := 0; i < 50; i++ {
			databaseNamesOperation := NewGetDatabaseNamesOperation(0, 20)
			command := databaseNamesOperation.getCommand(conventions)
			err := executor.executeCommand(command)
			assert.NoError(t, err)
		}
	}
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_canIssueManyRequests end\n")
	}
}

func requestExecutorTest_canFetchDatabasesNames(t *testing.T) {
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_canFetchDatabasesNames start\n")
	}
	conventions := NewDocumentConventions()
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		executor := RequestExecutor_create(store.getUrls(), store.getDatabase(), nil, conventions)

		databaseNamesOperation := NewGetDatabaseNamesOperation(0, 20)
		command := databaseNamesOperation.getCommand(conventions)
		err := executor.executeCommand(command)
		assert.NoError(t, err)

		dbNames := command.Result
		assert.True(t, stringArrayContains(dbNames, store.getDatabase()))
	}
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_canFetchDatabasesNames end\n")
	}
}

func requestExecutorTest_throwsWhenUpdatingTopologyOfNotExistingDb(t *testing.T) {
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_throwsWhenUpdatingTopologyOfNotExistingDb start\n")
	}
	conventions := NewDocumentConventions()
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		executor := RequestExecutor_create(store.getUrls(), "no_such_db", nil, conventions)
		serverNode := NewServerNode()
		serverNode.setUrl(store.getUrls()[0])
		serverNode.setDatabase("no_such")
		future := executor.updateTopologyAsync(serverNode, 5000)
		_, err := future.get()
		_ = err.(*DatabaseDoesNotExistException)
	}
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_throwsWhenUpdatingTopologyOfNotExistingDb end\n")
	}
}

func requestExecutorTest_throwsWhenDatabaseDoesNotExist(t *testing.T) {
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_throwsWhenDatabaseDoesNotExist start\n")
	}
	conventions := NewDocumentConventions()
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		executor := RequestExecutor_create(store.getUrls(), "no_such_db", nil, conventions)
		command := NewGetNextOperationIdCommand()
		err := executor.executeCommand(command)
		_ = err.(*DatabaseDoesNotExistException)
	}
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_throwsWhenDatabaseDoesNotExist end\n")
	}
}

func requestExecutorTest_canCreateSingleNodeRequestExecutor(t *testing.T) {
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_canCreateSingleNodeRequestExecutor start\n")
	}
	documentConventions := NewDocumentConventions()
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		executor := RequestExecutor_createForSingleNodeWithoutConfigurationUpdates(store.getUrls()[0], store.getDatabase(), nil, documentConventions)
		nodes := executor.getTopologyNodes()
		assert.Equal(t, 1, len(nodes))

		serverNode := nodes[0]
		assert.Equal(t, serverNode.getUrl(), store.getUrls()[0])
		assert.Equal(t, serverNode.getDatabase(), store.getDatabase())

		command := NewGetNextOperationIdCommand()
		err := executor.executeCommand(command)
		assert.NoError(t, err)
		assert.NotNil(t, command.Result)
	}
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_canCreateSingleNodeRequestExecutor end\n")
	}
}

func requestExecutorTest_canChooseOnlineNode(t *testing.T) {
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_canChooseOnlineNode start\n")
	}
	documentConventions := NewDocumentConventions()
	store := getDocumentStoreMust(t)
	defer store.Close()

	url := store.getUrls()[0]
	dbName := store.getDatabase()
	{
		executor := RequestExecutor_create([]string{"http://no_such_host:8080", "http://another_offlilne:8080", url}, dbName, nil, documentConventions)
		command := NewGetNextOperationIdCommand()
		err := executor.executeCommand(command)
		assert.NoError(t, err)
		assert.NotNil(t, command.Result)
		topologyNodes := executor.getTopologyNodes()
		assert.Equal(t, len(topologyNodes), 1)
		assert.Equal(t, url, topologyNodes[0].getUrl())
		assert.Equal(t, url, executor.getUrl())
	}
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_canChooseOnlineNode end\n")
	}
}

func requestExecutorTest_failsWhenServerIsOffline(t *testing.T) {
	if dbgRequestExecutorTests {
		logGoroutines("goroutines_req_executor_before.txt")
		fmt.Printf("requestExecutorTest_failsWhenServerIsOffline start\n")
	}
	documentConventions := NewDocumentConventions()
	executor := RequestExecutor_create([]string{"http://no_such_host:8081"}, "db1", nil, documentConventions)
	command := NewGetNextOperationIdCommand()
	err := executor.executeCommand(command)
	assert.Error(t, err)

	_ = err.(*AllTopologyNodesDownException)
}

func TestRequestExecutor(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	requestExecutorTest_canFetchDatabasesNames(t)
	requestExecutorTest_canIssueManyRequests(t)
	requestExecutorTest_throwsWhenDatabaseDoesNotExist(t)
	requestExecutorTest_failuresDoesNotBlockConnectionPool(t)
	requestExecutorTest_canCreateSingleNodeRequestExecutor(t)
	requestExecutorTest_failsWhenServerIsOffline(t)
	requestExecutorTest_throwsWhenUpdatingTopologyOfNotExistingDb(t)
	requestExecutorTest_canChooseOnlineNode(t)
}
