package main

import (
	"fmt"

	ravendb "github.com/ravendb/ravendb-go-client"
)

// stuff not yet implemented

func testGetClusterTopologyCommand2() {
	store := ravendb.NewDocumentStore([]string{serverURL}, "PyRavenDB")
	store.Initialize()
	re := store.GetRequestExecutor("")
	cmd := ravendb.NewGetClusterTopologyCommand()
	exec := re.GetCommandExecutor()
	rsp, err := ravendb.ExecuteGetClusterTopologyCommand(exec, cmd)
	must(err)
	fmt.Printf("%v\n", rsp)
}

func testDbNotExist() {
	store := ravendb.NewDocumentStore([]string{serverURL}, "PyRavenDB")
	store.Initialize()
	session, err := store.OpenSession()
	must(err)
	fmt.Printf("session: %v\n", session)
}
