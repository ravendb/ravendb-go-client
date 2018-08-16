package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

func getClusterTopologyTest_canGetTopology(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	command := ravendb.NewGetClusterTopologyCommand()
	err = store.GetRequestExecutor().ExecuteCommand(command)
	assert.NoError(t, err)
	result := command.Result
	assert.NotNil(t, result)

	assert.NotEmpty(t, result.GetLeader())
	assert.NotEmpty(t, result.GetNodeTag())

	topology := result.GetTopology()
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

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	getClusterTopologyTest_canGetTopology(t)
}
