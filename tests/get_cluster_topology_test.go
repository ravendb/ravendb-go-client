package tests

import (
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func getClusterTopologyTestCanGetTopology(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	command := ravendb.NewGetClusterTopologyCommand()
	err = store.GetRequestExecutor("").ExecuteCommand(command, nil)
	assert.NoError(t, err)
	result := command.Result
	assert.NotNil(t, result)

	assert.NotEmpty(t, result.Leader)
	assert.NotEmpty(t, result.NodeTag)

	topology := result.Topology
	assert.NotNil(t, topology)
	assert.NotEmpty(t, topology.TopologyID)
	assert.Equal(t, 1, len(topology.Members))
	assert.Equal(t, 0, len(topology.Watchers))
	assert.Equal(t, 0, len(topology.Promotables))
}

func TestGetClusterTopology(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	getClusterTopologyTestCanGetTopology(t, driver)
}
