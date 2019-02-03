package tests

import (
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func getTopologyTestCanGetTopology(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	command := ravendb.NewGetDatabaseTopologyCommand()
	err = store.GetRequestExecutor("").ExecuteCommand(command, nil)
	assert.NoError(t, err)
	result := command.Result
	assert.NotNil(t, result)

	assert.NotEqual(t, result.Etag, "")
	assert.Equal(t, len(result.Nodes), 1)
	serverNode := result.Nodes[0]
	assert.Equal(t, serverNode.URL, store.GetUrls()[0])
	assert.Equal(t, serverNode.Database, store.GetDatabase())
	assert.Equal(t, serverNode.ClusterTag, "A")
	assert.Equal(t, serverNode.ServerRole, ravendb.ServerNodeRoleMember)
}

func TestGetTopology(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	getTopologyTestCanGetTopology(t, driver)
}
