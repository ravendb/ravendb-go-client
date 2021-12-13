package tests

import (
	"github.com/ravendb/ravendb-go-client/serverwide/operations"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func removeNodeFromClusterTest(t *testing.T, driver *RavenTestDriver) {
	if os.Getenv("RAVEN_License") == "" {
		t.Skip("This test requires RavenDB license.")
	}
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()
	assert.NoError(t, err)

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

	time.Sleep(time.Second * 5) // wait for topology to be updated
	operation := operations.OperationGetClusterTopology{}
	err = store.Maintenance().Server().Send(&operation)
	assert.NoError(t, err)
	assert.NotEmpty(t, operation.Leader)
	assert.NotEmpty(t, operation.NodeTag)

	topology := operation.Topology
	assert.NotNil(t, topology)
	assert.NotEmpty(t, topology.TopologyID)
	assert.Equal(t, 2, len(topology.Members))
	assert.Equal(t, 0, len(topology.Watchers))
	assert.Equal(t, 0, len(topology.Promotables))

	operationRemoveNode := operations.RemoveClusterNode{
		Node: store2.GetUrls()[0],
		Tag:  "B",
	}
	err = store.Maintenance().Server().Send(&operationRemoveNode)
	assert.NoError(t, err)

	time.Sleep(time.Second * 5) // wait for topology to be updated
	operation = operations.OperationGetClusterTopology{}
	err = store.Maintenance().Server().Send(&operation)
	assert.NoError(t, err)
	assert.NotEmpty(t, operation.Leader)
	assert.NotEmpty(t, operation.NodeTag)

	topology = operation.Topology
	assert.NotNil(t, topology)
	assert.NotEmpty(t, topology.TopologyID)
	assert.Equal(t, 1, len(topology.Members))
	assert.Equal(t, 0, len(topology.Watchers))
	assert.Equal(t, 0, len(topology.Promotables))
}

func TestRemoveNodeFromCluster(t *testing.T) {
	driver := createTestDriver(t)

	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	removeNodeFromClusterTest(t, driver)
}
