package ravendb

import (
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

func getClusterTopologyTest_canGetTopology(t *testing.T) {
	store := getDocumentStoreMust(t)
	command := NewGetClusterTopologyCommand()
	err := store.GetRequestExecutor().executeCommand(command)
	assert.NoError(t, err)
	result := command.Result
	assert.NotNil(t, result)

	assert.NotEmpty(t, result.getLeader())
	assert.NotEmpty(t, result.getNodeTag())

	topology := result.getTopology()
	assert.NotNil(t, topology)
	assert.NotEmpty(t, topology.TopologyID)
	assert.Equal(t, 1, len(topology.Members))
	assert.Equal(t, 0, len(topology.Watchers))
	assert.Equal(t, 0, len(topology.Promotables))
}

func TestGetClusterTopology(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	if useProxy() {
		proxy.ChangeLogFile("trace_get_cluster_topology_go.txt")
	}

	createTestDriver()
	defer deleteTestDriver()

	getClusterTopologyTest_canGetTopology(t)
}
