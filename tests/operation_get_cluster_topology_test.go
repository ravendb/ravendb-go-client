package tests

import (
	"github.com/ravendb/ravendb-go-client/serverwide/operations"
	"github.com/stretchr/testify/assert"
	"testing"
)

func getClusterTopologyTestCanGetTopology(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()
	operation := operations.OperationGetClusterTopology{}
	err = store.Maintenance().Server().Send(&operation)
	assert.NoError(t, err)

	assert.NotEmpty(t, operation.Leader)
	assert.NotEmpty(t, operation.NodeTag)

	topology := operation.Topology
	assert.NotNil(t, topology)
	assert.NotEmpty(t, topology.TopologyID)
	assert.Equal(t, 1, len(topology.Members))
	assert.Equal(t, 0, len(topology.Watchers))
	assert.Equal(t, 0, len(topology.Promotables))
}

func TestGetClusterTopology(t *testing.T) {
    	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	getClusterTopologyTestCanGetTopology(t, driver)
}
