package main

import (
	"fmt"

	ravendb "github.com/ravendb/ravendb-go-client"
)

var (
	dbName = "Demo"
)

func panicIfErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func getDocumentStore(databaseName string) (*ravendb.DocumentStore, error) {
	serverNodes := []string{"http://live-test.ravendb.net"}
	store := ravendb.NewDocumentStore(serverNodes, databaseName)
	if err := store.Initialize(); err != nil {
		return nil, err
	}
	return store, nil
}

func loadUpdateSave() {
	/*
		store, err := getDocumentStore(dbName)
		panicIfErr(err)
	*/
}

func loadEmployee() {
	store, err := getDocumentStore(dbName)
	panicIfErr(err)

	session, err := store.OpenSession("")
	panicIfErr(err)
	defer session.Close()
	var e *Employee
	err = session.Load(&e, "employees/7-A")
	panicIfErr(err)
	fmt.Printf("employee: %#v\n", e)
	session.Close()

	store.Close()
}

func loadWithIncludes() {
	store, err := getDocumentStore(dbName)
	panicIfErr(err)

	{
		// TODO: not working yet, see https://github.com/ravendb/ravendb-go-client/issues/63
		session, err := store.OpenSession("")
		panicIfErr(err)
		var o *Order
		err = session.Include("employee").Load(&o, "orders/827-A")
		if err != nil {
			fmt.Printf("error: %s\n", err)
		} else {
			fmt.Printf("order: %#v\n", o)
		}
	}

}

func main() {
}
