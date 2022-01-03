package tests

import (
	"errors"
	"github.com/ravendb/ravendb-go-client"
	"github.com/ravendb/ravendb-go-client/serverwide/operations"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

const iterations = 10

func AddDatabaseToNode(t *testing.T, driver *RavenTestDriver) error {
	if os.Getenv("RAVEN_License") == "" {
		t.Skip("This test requires RavenDB license.")
	}
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	driver2 := createTestDriver(t)
	store2, err := driver2.createMainStore()
	defer store2.Close()
	assert.NoError(t, err)

	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	operationAddNodeToCluster := operations.OperationAddClusterNode{
		Url:     store2.GetUrls()[0],
		Tag:     "B",
		Watcher: false,
	}

	err = store.Maintenance().Server().Send(&operationAddNodeToCluster)
	assert.NoError(t, err)

	operationAddDatabaseNode := operations.OperationAddDatabaseNode{
		Name: store.GetDatabase(),
		Node: "B",
	}
	err = store.Maintenance().Server().Send(&operationAddDatabaseNode)
	assert.NoError(t, err)

	for i := 0; i <= iterations; i++ {
		time.Sleep(2 * time.Second) //we HAVE to wait for things to move around
		command := ravendb.NewGetDatabaseTopologyCommand()
		err = store2.GetRequestExecutor(store.GetDatabase()).ExecuteCommand(command, nil)
		assert.NoError(t, err)
		result := command.Result
		assert.NotNil(t, result)
		if len(result.Nodes) == 2 {
			assert.Equal(t, result.Nodes[0].Database, store.GetDatabase())
			assert.Equal(t, result.Nodes[1].Database, store.GetDatabase())
			break
		} else if i == iterations {
			return errors.New("Expected database to be included in both node A and node B. ")
		}
	}
	return nil
}

func TestAddDatabaseToNode(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	err := AddDatabaseToNode(t, driver)
	if err != nil {
		panic(err.Error())
	}
}
