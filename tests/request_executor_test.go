package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

var (
	dbgRequestExecutorTests = false
)

func requestExecutorTest_failuresDoesNotBlockConnectionPool(t *testing.T) {
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_failuresDoesNotBlockConnectionPool start\n")
	}
	conventions := ravendb.NewDocumentConventions()
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		executor := ravendb.RequestExecutor_create(store.GetUrls(), "no_such_db", nil, conventions)
		errorsCount := 0

		for i := 0; i < 40; i++ {
			command := ravendb.NewGetNextOperationIdCommand()
			err := executor.ExecuteCommand(command)
			if err != nil {
				errorsCount++
			}
		}
		assert.Equal(t, 40, errorsCount)

		databaseNamesOperation := ravendb.NewGetDatabaseNamesOperation(0, 20)
		command := databaseNamesOperation.GetCommand(conventions)
		err := executor.ExecuteCommand(command)
		_ = err.(*ravendb.DatabaseDoesNotExistException)
	}
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_failuresDoesNotBlockConnectionPool end\n")
	}
}

func requestExecutorTest_canIssueManyRequests(t *testing.T) {
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_canIssueManyRequests start\n")
	}
	conventions := ravendb.NewDocumentConventions()
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		executor := ravendb.RequestExecutor_create(store.GetUrls(), store.GetDatabase(), nil, conventions)
		for i := 0; i < 50; i++ {
			databaseNamesOperation := ravendb.NewGetDatabaseNamesOperation(0, 20)
			command := databaseNamesOperation.GetCommand(conventions)
			err := executor.ExecuteCommand(command)
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
	conventions := ravendb.NewDocumentConventions()
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		executor := ravendb.RequestExecutor_create(store.GetUrls(), store.GetDatabase(), nil, conventions)

		databaseNamesOperation := ravendb.NewGetDatabaseNamesOperation(0, 20)
		command := databaseNamesOperation.GetCommand(conventions)
		err := executor.ExecuteCommand(command)
		assert.NoError(t, err)

		dbNames := command.Result
		assert.True(t, ravendb.StringArrayContains(dbNames, store.GetDatabase()))
	}
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_canFetchDatabasesNames end\n")
	}
}

func requestExecutorTest_throwsWhenUpdatingTopologyOfNotExistingDb(t *testing.T) {
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_throwsWhenUpdatingTopologyOfNotExistingDb start\n")
	}
	conventions := ravendb.NewDocumentConventions()
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		executor := ravendb.RequestExecutor_create(store.GetUrls(), "no_such_db", nil, conventions)
		serverNode := ravendb.NewServerNode()
		serverNode.SetUrl(store.GetUrls()[0])
		serverNode.SetDatabase("no_such")
		future := executor.UpdateTopologyAsync(serverNode, 5000)
		_, err := future.Get()
		_ = err.(*ravendb.DatabaseDoesNotExistException)
	}
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_throwsWhenUpdatingTopologyOfNotExistingDb end\n")
	}
}

func requestExecutorTest_throwsWhenDatabaseDoesNotExist(t *testing.T) {
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_throwsWhenDatabaseDoesNotExist start\n")
	}
	conventions := ravendb.NewDocumentConventions()
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		executor := ravendb.RequestExecutor_create(store.GetUrls(), "no_such_db", nil, conventions)
		command := ravendb.NewGetNextOperationIdCommand()
		err := executor.ExecuteCommand(command)
		_ = err.(*ravendb.DatabaseDoesNotExistException)
	}
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_throwsWhenDatabaseDoesNotExist end\n")
	}
}

func requestExecutorTest_canCreateSingleNodeRequestExecutor(t *testing.T) {
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_canCreateSingleNodeRequestExecutor start\n")
	}
	documentConventions := ravendb.NewDocumentConventions()
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		executor :=ravendb. RequestExecutor_createForSingleNodeWithoutConfigurationUpdates(store.GetUrls()[0], store.GetDatabase(), nil, documentConventions)
		nodes := executor.GetTopologyNodes()
		assert.Equal(t, 1, len(nodes))

		serverNode := nodes[0]
		assert.Equal(t, serverNode.GetUrl(), store.GetUrls()[0])
		assert.Equal(t, serverNode.GetDatabase(), store.GetDatabase())

		command := ravendb.NewGetNextOperationIdCommand()
		err := executor.ExecuteCommand(command)
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
	documentConventions := ravendb.NewDocumentConventions()
	store := getDocumentStoreMust(t)
	defer store.Close()

	url := store.GetUrls()[0]
	dbName := store.GetDatabase()
	{
		executor := ravendb.RequestExecutor_create([]string{"http://no_such_host:8080", "http://another_offlilne:8080", url}, dbName, nil, documentConventions)
		command := ravendb.NewGetNextOperationIdCommand()
		err := executor.ExecuteCommand(command)
		assert.NoError(t, err)
		assert.NotNil(t, command.Result)
		topologyNodes := executor.GetTopologyNodes()
		assert.Equal(t, len(topologyNodes), 1)
		assert.Equal(t, url, topologyNodes[0].GetUrl())
		assert.Equal(t, url, executor.GetUrl())
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
	documentConventions := ravendb.NewDocumentConventions()
	executor := ravendb.RequestExecutor_create([]string{"http://no_such_host:8081"}, "db1", nil, documentConventions)
	command := ravendb.NewGetNextOperationIdCommand()
	err := executor.ExecuteCommand(command)
	assert.Error(t, err)

	_ = err.(*ravendb.AllTopologyNodesDownException)
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
