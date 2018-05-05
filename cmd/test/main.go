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

func testGetDatabaseNamesCommand() {
	exec := getExecutor()
	cmd := ravendb.NewGetDatabaseNamesCommand(0, 3)
	res, err := ravendb.ExecuteGetDatabaseNamesCommand(exec, cmd, false)
	must(err)
	if verboseLog {
		fmt.Printf("databases: %#v\n", res.Databases)
	}
	fmt.Printf("testGetDatabaseNamesCommand ok\n")
}

func testCreateDatabaseCommand() {
	exec := getExecutor()
	dbName := ravendb.NewUUID().Hex()
	cmd := ravendb.NewCreateDatabaseCommand(dbName, 1)
	res, err := ravendb.ExecuteCreateDatabaseCommand(exec, cmd, false)
	must(err)
	panicIf(res.RaftCommandIndex == 0, "res.RaftCommandIndex is 0")
	panicIf(res.Name != dbName, "res.Name is '%s', expected '%s'", res.Name, dbName)
	if verboseLog {
		fmt.Printf("res: %#v\n", res)
	}
	fmt.Printf("testCreateDatabaseCommand ok\n")
}

func testCreateAndDeleteDatabaseCommand() {
	dbName := ravendb.NewUUID().Hex()
	exec := getExecutor()
	cmd := ravendb.NewCreateDatabaseCommand(dbName, 1)
	res, err := ravendb.ExecuteCreateDatabaseCommand(exec, cmd, false)
	must(err)
	panicIf(res.RaftCommandIndex == 0, "res.RaftCommandIndex is 0")
	panicIf(res.Name != dbName, "res.Name is '%s', expected '%s'", res.Name, dbName)
	if true || verboseLog {
		fmt.Printf("res: %#v\n", res)
	}

	// TODO: do I need to wait?

	cmd2 := ravendb.NewDeleteDatabaseCommand(dbName, false, "")
	res2, err := ravendb.ExecuteDeleteDatabaseCommand(exec, cmd2, false)
	must(err)
	panicIf(res2.RaftCommandIndex == 0, "res2.RaftCommandIndex is 0")
	if verboseLog {
		fmt.Printf("res2: %#v\n", res2)
	}

	fmt.Printf("testCreateAndDeleteDatabaseCommand ok\n")
}

func main() {
	//testInvalidCommand()
	//testGetClusterTopologyCommand()
	//testGetStatisticsCommand()
	//testGetStatisticsCommandBadDb()
	//testGetTopologyCommand()
	//testGetTopologyCommandBadDb()
	//testGetDatabaseNamesCommand()
	//testCreateDatabaseCommand()
	testCreateAndDeleteDatabaseCommand()
}
