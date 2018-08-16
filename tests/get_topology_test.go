package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

func getTopologyTest_canGetTopology(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	command := ravendb.NewGetDatabaseTopologyCommand()
	err = store.GetRequestExecutor().ExecuteCommand(command)
	assert.NoError(t, err)
	result := command.Result
	assert.NotNil(t, result)

	assert.NotEqual(t, result.GetEtag(), "")
	assert.Equal(t, len(result.GetNodes()), 1)
	serverNode := result.GetNodes()[0]
	assert.Equal(t, serverNode.GetUrl(), store.GetUrls()[0])
	assert.Equal(t, serverNode.GetDatabase(), store.GetDatabase())
	assert.Equal(t, serverNode.GetClusterTag(), "A")
	assert.Equal(t, serverNode.GetServerRole(), ravendb.ServerNode_Role_MEMBER)
}

func TestGetTopology(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	getTopologyTest_canGetTopology(t)
}
