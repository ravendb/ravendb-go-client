package tests

import (
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func requestExecutorTestFailuresDoesNotBlockConnectionPool(t *testing.T, driver *RavenTestDriver) {
	conventions := ravendb.NewDocumentConventions()
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		executor := ravendb.RequestExecutorCreate(store.GetUrls(), "no_such_db", nil, nil, conventions)
		errorsCount := 0

		for i := 0; i < 40; i++ {
			command := ravendb.NewGetNextOperationIDCommand()
			err := executor.ExecuteCommand(command, nil)
			if err != nil {
				errorsCount++
			}
		}
		assert.Equal(t, 40, errorsCount)

		databaseNamesOperation := ravendb.NewGetDatabaseNamesOperation(0, 20)
		command := databaseNamesOperation.GetCommand(conventions)
		err := executor.ExecuteCommand(command, nil)
		_ = err.(*ravendb.DatabaseDoesNotExistError)
	}
}

func requestExecutorTestCanIssueManyRequests(t *testing.T, driver *RavenTestDriver) {
	conventions := ravendb.NewDocumentConventions()
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		executor := ravendb.RequestExecutorCreate(store.GetUrls(), store.GetDatabase(), nil, nil, conventions)
		for i := 0; i < 50; i++ {
			databaseNamesOperation := ravendb.NewGetDatabaseNamesOperation(0, 20)
			command := databaseNamesOperation.GetCommand(conventions)
			err := executor.ExecuteCommand(command, nil)
			assert.NoError(t, err)
		}
	}
}

func requestExecutorTestCanFetchDatabasesNames(t *testing.T, driver *RavenTestDriver) {
	conventions := ravendb.NewDocumentConventions()
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		executor := ravendb.RequestExecutorCreate(store.GetUrls(), store.GetDatabase(), nil, nil, conventions)

		databaseNamesOperation := ravendb.NewGetDatabaseNamesOperation(0, 20)
		command := databaseNamesOperation.GetCommand(conventions)
		err := executor.ExecuteCommand(command, nil)
		assert.NoError(t, err)

		dbNames := command.Result
		assert.True(t, stringArrayContains(dbNames, store.GetDatabase()))
	}
}

func requestExecutorTestThrowsWhenUpdatingTopologyOfNotExistingDb(t *testing.T, driver *RavenTestDriver) {
	conventions := ravendb.NewDocumentConventions()
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		executor := ravendb.RequestExecutorCreate(store.GetUrls(), "no_such_db", nil, nil, conventions)
		serverNode := ravendb.NewServerNode()
		serverNode.URL = store.GetUrls()[0]
		serverNode.Database = "no_such"
		future := executor.UpdateTopologyAsync(serverNode, 5000)
		result := <-future
		err := result.Err
		_ = err.(*ravendb.DatabaseDoesNotExistError)
	}
}

func requestExecutorTestThrowsWhenDatabaseDoesNotExist(t *testing.T, driver *RavenTestDriver) {
	conventions := ravendb.NewDocumentConventions()
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		executor := ravendb.RequestExecutorCreate(store.GetUrls(), "no_such_db", nil, nil, conventions)
		command := ravendb.NewGetNextOperationIDCommand()
		err := executor.ExecuteCommand(command, nil)
		_ = err.(*ravendb.DatabaseDoesNotExistError)
	}
}

func requestExecutorTestCanCreateSingleNodeRequestExecutor(t *testing.T, driver *RavenTestDriver) {
	documentConventions := ravendb.NewDocumentConventions()
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		executor := ravendb.RequestExecutorCreateForSingleNodeWithoutConfigurationUpdates(store.GetUrls()[0], store.GetDatabase(), nil, nil, documentConventions)
		nodes := executor.GetTopologyNodes()
		assert.Equal(t, 1, len(nodes))

		serverNode := nodes[0]
		assert.Equal(t, serverNode.URL, store.GetUrls()[0])
		assert.Equal(t, serverNode.Database, store.GetDatabase())

		command := ravendb.NewGetNextOperationIDCommand()
		err := executor.ExecuteCommand(command, nil)
		assert.NoError(t, err)
		assert.NotNil(t, command.Result)
	}
}

func requestExecutorTestCanChooseOnlineNode(t *testing.T, driver *RavenTestDriver) {
	documentConventions := ravendb.NewDocumentConventions()
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	url := store.GetUrls()[0]
	dbName := store.GetDatabase()
	{
		executor := ravendb.RequestExecutorCreate([]string{"http://no_such_host:8080", "http://another_offlilne:8080", url}, dbName, nil, nil, documentConventions)
		command := ravendb.NewGetNextOperationIDCommand()
		err := executor.ExecuteCommand(command, nil)
		assert.NoError(t, err)
		assert.NotNil(t, command.Result)
		topologyNodes := executor.GetTopologyNodes()
		assert.Equal(t, len(topologyNodes), 1)
		assert.Equal(t, url, topologyNodes[0].URL)
		url2, err := executor.GetURL()
		assert.NoError(t, err)
		assert.Equal(t, url, url2)
	}
}

func requestExecutorTestFailsWhenServerIsOffline(t *testing.T) {
	documentConventions := ravendb.NewDocumentConventions()
	executor := ravendb.RequestExecutorCreate([]string{"http://no_such_host:8081"}, "db1", nil, nil, documentConventions)
	command := ravendb.NewGetNextOperationIDCommand()
	err := executor.ExecuteCommand(command, nil)
	assert.Error(t, err)

	_ = err.(*ravendb.AllTopologyNodesDownError)
}

func TestRequestExecutor(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	requestExecutorTestCanFetchDatabasesNames(t, driver)
	requestExecutorTestCanIssueManyRequests(t, driver)
	requestExecutorTestThrowsWhenDatabaseDoesNotExist(t, driver)
	requestExecutorTestFailuresDoesNotBlockConnectionPool(t, driver)
	requestExecutorTestCanCreateSingleNodeRequestExecutor(t, driver)
	requestExecutorTestFailsWhenServerIsOffline(t)
	requestExecutorTestThrowsWhenUpdatingTopologyOfNotExistingDb(t, driver)
	requestExecutorTestCanChooseOnlineNode(t, driver)
}
