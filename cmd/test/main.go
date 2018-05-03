package main

import (
	"fmt"

	ravendb "github.com/ravendb/ravendb-go-client"
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func testDbNotExist() {
	store := ravendb.NewDocumentStore([]string{"http://localhost:9999"}, "PyRavenDB")
	store.Initialize()
	session, err := store.OpenSession()
	must(err)
	fmt.Printf("session: %v\n", session)
}

func main() {
	var cmd ravendb.RavenCommand
	fmt.Printf("cmd: %v\n", cmd)
}
