package main

import (
	"fmt"
	"net/http"

	ravendb "github.com/ravendb/ravendb-go-client"
)

var (
	serverURL = "http://localhost:9999"
	dbName    = "PyRavenDB"
)

func getExecutor() ravendb.CommandExecutorFunc {
	node := &ravendb.ServerNode{
		URL:        serverURL,
		Database:   dbName,
		ClusterTag: "0",
	}
	return ravendb.MakeSimpleExecutor(node)
}

// test that when we send invalid command to the server, we get the right
// error code
func testInvalidCommand() {
	exec := getExecutor()
	cmd := &ravendb.RavenCommand{
		Method:        http.MethodGet,
		IsReadRequest: true,
		URLTemplate:   "{url}/cluster/invalid",
	}
	clusterTopology, err := ravendb.ExecuteGetClusterTopologyCommand(exec, cmd, false)
	panicIf(clusterTopology != nil, "expected nil clusterTopology")
	re := err.(*ravendb.BadRequestError)
	panicIf(re.URL != "/cluster/invalid", "unexpected re.URL. is '%s', should be '/cluster/invalid'", re.URL)
	fmt.Printf("testInvalidCommand ok\n")
}

func testGetClusterTopologyCommand() {
	exec := getExecutor()
	cmd := ravendb.NewGetClusterTopologyCommand()
	clusterTopology, err := ravendb.ExecuteGetClusterTopologyCommand(exec, cmd, false)
	must(err)
	fmt.Printf("  %#v\n", clusterTopology)
	fmt.Printf("  %#v\n", clusterTopology.Topology)
	fmt.Printf("testGetClusterTopologyCommand ok\n")
}

func main() {
	testInvalidCommand()
	testGetClusterTopologyCommand()
}
