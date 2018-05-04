package main

import (
	"fmt"
	"net/http"

	ravendb "github.com/ravendb/ravendb-go-client"
)

var (
	serverURL = "http://localhost:9999"
	dbName    = "PyRavenDB"

	// enable to see more information for each test
	verboseLog = false
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
	nServers := len(clusterTopology.Topology.Members)
	panicIf(nServers < 1, "returned no Members server, expected at least 1")
	// Note: not sure if the name will always be "A", that's what happens when
	// I run agains my local setup
	panicIf(clusterTopology.Leader != "A", "clusterTopology.Leader is '%s', expected 'A'", clusterTopology.Leader)
	if verboseLog {
		fmt.Printf("  %#v\n", clusterTopology)
		fmt.Printf("  %#v\n", clusterTopology.Topology)
	}
	fmt.Printf("testGetClusterTopologyCommand ok\n")
}

func main() {
	testInvalidCommand()
	testGetClusterTopologyCommand()
}
