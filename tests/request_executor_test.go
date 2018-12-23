package tests

import (
	"fmt"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

var (
	dbgRequestExecutorTests = false
)

func requestExecutorTest_failuresDoesNotBlockConnectionPool(t *testing.T, driver *RavenTestDriver) {
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_failuresDoesNotBlockConnectionPool start\n")
	}
	conventions := ravendb.NewDocumentConventions()
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		executor := ravendb.RequestExecutor_create(store.GetUrls(), "no_such_db", nil, conventions)
		errorsCount := 0

		for i := 0; i < 40; i++ {
			command := ravendb.NewGetNextOperationIDCommand()
			err := executor.ExecuteCommand(command)
			if err != nil {
				errorsCount++
			}
		}
		assert.Equal(t, 40, errorsCount)

		databaseNamesOperation := ravendb.NewGetDatabaseNamesOperation(0, 20)
		command := databaseNamesOperation.GetCommand(conventions)
		err := executor.ExecuteCommand(command)
		_ = err.(*ravendb.DatabaseDoesNotExistError)
	}
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_failuresDoesNotBlockConnectionPool end\n")
	}
}

func requestExecutorTest_canIssueManyRequests(t *testing.T, driver *RavenTestDriver) {
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_canIssueManyRequests start\n")
	}
	conventions := ravendb.NewDocumentConventions()
	store := getDocumentStoreMust(t, driver)
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

func requestExecutorTest_canFetchDatabasesNames(t *testing.T, driver *RavenTestDriver) {
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_canFetchDatabasesNames start\n")
	}
	conventions := ravendb.NewDocumentConventions()
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		executor := ravendb.RequestExecutor_create(store.GetUrls(), store.GetDatabase(), nil, conventions)

		databaseNamesOperation := ravendb.NewGetDatabaseNamesOperation(0, 20)
		command := databaseNamesOperation.GetCommand(conventions)
		err := executor.ExecuteCommand(command)
		assert.NoError(t, err)

		dbNames := command.Result
		assert.True(t, stringArrayContains(dbNames, store.GetDatabase()))
	}
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_canFetchDatabasesNames end\n")
	}
}

func requestExecutorTest_throwsWhenUpdatingTopologyOfNotExistingDb(t *testing.T, driver *RavenTestDriver) {
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_throwsWhenUpdatingTopologyOfNotExistingDb start\n")
	}
	conventions := ravendb.NewDocumentConventions()
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		executor := ravendb.RequestExecutor_create(store.GetUrls(), "no_such_db", nil, conventions)
		serverNode := ravendb.NewServerNode()
		serverNode.SetUrl(store.GetUrls()[0])
		serverNode.SetDatabase("no_such")
		future := executor.UpdateTopologyAsync(serverNode, 5000)
		_, err := future.Get()
		_ = err.(*ravendb.DatabaseDoesNotExistError)
	}
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_throwsWhenUpdatingTopologyOfNotExistingDb end\n")
	}
}

func requestExecutorTest_throwsWhenDatabaseDoesNotExist(t *testing.T, driver *RavenTestDriver) {
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_throwsWhenDatabaseDoesNotExist start\n")
	}
	conventions := ravendb.NewDocumentConventions()
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		executor := ravendb.RequestExecutor_create(store.GetUrls(), "no_such_db", nil, conventions)
		command := ravendb.NewGetNextOperationIDCommand()
		err := executor.ExecuteCommand(command)
		_ = err.(*ravendb.DatabaseDoesNotExistError)
	}
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_throwsWhenDatabaseDoesNotExist end\n")
	}
}

func requestExecutorTest_canCreateSingleNodeRequestExecutor(t *testing.T, driver *RavenTestDriver) {
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_canCreateSingleNodeRequestExecutor start\n")
	}
	documentConventions := ravendb.NewDocumentConventions()
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		executor := ravendb.RequestExecutor_createForSingleNodeWithoutConfigurationUpdates(store.GetUrls()[0], store.GetDatabase(), nil, documentConventions)
		nodes := executor.GetTopologyNodes()
		assert.Equal(t, 1, len(nodes))

		serverNode := nodes[0]
		assert.Equal(t, serverNode.GetUrl(), store.GetUrls()[0])
		assert.Equal(t, serverNode.GetDatabase(), store.GetDatabase())

		command := ravendb.NewGetNextOperationIDCommand()
		err := executor.ExecuteCommand(command)
		assert.NoError(t, err)
		assert.NotNil(t, command.Result)
	}
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_canCreateSingleNodeRequestExecutor end\n")
	}
}

func requestExecutorTest_canChooseOnlineNode(t *testing.T, driver *RavenTestDriver) {
	if dbgRequestExecutorTests {
		fmt.Printf("requestExecutorTest_canChooseOnlineNode start\n")
	}
	documentConventions := ravendb.NewDocumentConventions()
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	url := store.GetUrls()[0]
	dbName := store.GetDatabase()
	{
		executor := ravendb.RequestExecutor_create([]string{"http://no_such_host:8080", "http://another_offlilne:8080", url}, dbName, nil, documentConventions)
		command := ravendb.NewGetNextOperationIDCommand()
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

func requestExecutorTest_failsWhenServerIsOffline(t *testing.T, driver *RavenTestDriver) {
	if dbgRequestExecutorTests {
		logGoroutines("goroutines_req_executor_before.txt")
		fmt.Printf("requestExecutorTest_failsWhenServerIsOffline start\n")
	}
	documentConventions := ravendb.NewDocumentConventions()
	executor := ravendb.RequestExecutor_create([]string{"http://no_such_host:8081"}, "db1", nil, documentConventions)
	command := ravendb.NewGetNextOperationIDCommand()
	err := executor.ExecuteCommand(command)
	assert.Error(t, err)

	_ = err.(*ravendb.AllTopologyNodesDownError)
}

func TestRequestExecutor(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	requestExecutorTest_canFetchDatabasesNames(t, driver)
	requestExecutorTest_canIssueManyRequests(t, driver)
	requestExecutorTest_throwsWhenDatabaseDoesNotExist(t, driver)
	requestExecutorTest_failuresDoesNotBlockConnectionPool(t, driver)
	requestExecutorTest_canCreateSingleNodeRequestExecutor(t, driver)
	requestExecutorTest_failsWhenServerIsOffline(t, driver)
	requestExecutorTest_throwsWhenUpdatingTopologyOfNotExistingDb(t, driver)
	requestExecutorTest_canChooseOnlineNode(t, driver)
}
