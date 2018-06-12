package ravendb

import (
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

func GetTopologyTest_canGetTopology(t *testing.T) {
	store := getDocumentStoreMust(t)
	command := NewGetDatabaseTopologyCommand()
	err := store.GetRequestExecutor().executeCommand(command)
	assert.NoError(t, err)
	result := command.Result
	assert.NotNil(t, result)

	assert.NotEqual(t, result.getEtag(), "")
	assert.Equal(t, len(result.getNodes()), 1)
	serverNode := result.getNodes()[0]
	assert.Equal(t, serverNode.getUrl(), store.getUrls()[0])
	assert.Equal(t, serverNode.getDatabase(), store.getDatabase())
	assert.Equal(t, serverNode.getClusterTag(), "A")
	assert.Equal(t, serverNode.getServerRole(), ServerNode_Role_MEMBER)
}

func TestGetTopology(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	if useProxy() {
		proxy.ChangeLogFile("trace_get_topology_go.txt")
	}
	GetTopologyTest_canGetTopology(t)
}
