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

func getInvalidDbExecutor() ravendb.CommandExecutorFunc {
	node := &ravendb.ServerNode{
		URL:        serverURL,
		Database:   "invalid-database",
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

func testGetStatisticsCommand() {
	exec := getExecutor()
	cmd := ravendb.NewGetStatisticsCommand("")
	stats, err := ravendb.ExecuteGetStatisticsCommand(exec, cmd, false)
	must(err)
	if verboseLog {
		fmt.Printf("stats: %#v\n", stats)
	}
	fmt.Printf("testGetStatisticsCommand ok\n")
}

func testGetStatisticsCommandBadDb() {
	exec := getInvalidDbExecutor()
	cmd := ravendb.NewGetStatisticsCommand("")
	res, err := ravendb.ExecuteGetStatisticsCommand(exec, cmd, false)
	panicIf(res != nil, "expected res to be nil")
	// TODO: should this be 501? In Python test it's not possible to execute
	// this command directly, it'll fail after GetTopology command
	re := err.(*ravendb.InternalServerError)
	if verboseLog {
		fmt.Printf("error: %s\n", re)
	}
	fmt.Printf("testGetStatisticsCommandBadDb ok\n")
}

func testGetTopologyCommand() {
	exec := getExecutor()
	cmd := ravendb.NewGetTopologyCommand()
	res, err := ravendb.ExecuteGetTopologyCommand(exec, cmd, false)
	must(err)
	if verboseLog {
		fmt.Printf("topology: %#v\n", res)
	}
	fmt.Printf("testGetTopologyCommand ok\n")
}

func testGetTopologyCommandBadDb() {
	exec := getInvalidDbExecutor()
	cmd := ravendb.NewGetTopologyCommand()
	res, err := ravendb.ExecuteGetTopologyCommand(exec, cmd, false)
	panicIf(res != nil, "expected res to be nil")
	panicIf(err == nil, "expected err to be non nil")
	re := err.(*ravendb.ServiceUnavailableError)
	if verboseLog {
		fmt.Printf("error: %s\n", re)
	}
	fmt.Printf("testGetTopologyCommandBadDb ok\n")
}

func main() {
	//testInvalidCommand()
	//testGetClusterTopologyCommand()
	//testGetStatisticsCommand()
	//testGetStatisticsCommandBadDb()
	//testGetTopologyCommand()
	testGetTopologyCommandBadDb()
}
