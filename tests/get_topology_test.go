package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

func getTopologyTest_canGetTopology(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
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
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	getTopologyTest_canGetTopology(t, driver)
}
