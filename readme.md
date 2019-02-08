[![Linux build Status](https://travis-ci.org/ravendb/ravendb-go-client.svg?branch=master)](https://travis-ci.org/ravendb/ravendb-go-client) [![Windows build status](https://ci.appveyor.com/api/projects/status/rf326yoxl1uf444h/branch/master?svg=true)](https://ci.appveyor.com/project/ravendb/ravendb-go-client/branch/master)

This is information on how to use the library. For docs on working on the library itself see [readme-dev.md](readme-dev.md).

This library requires go 1.11 or later.

Godoc: https://godoc.org/github.com/ravendb/ravendb-go-client

## Documentation

Please find the official documentation on [RavenDB Documentation](https://ravendb.net/docs/article-page/4.0/nodejs/client-api/what-is-a-document-store) page

## Getting started

Full source code of those examples is in `examples` directory.

1. Import the package
```go
import (
	ravendb "github.com/ravendb/ravendb-go-client"
)
```
2. Initialize document store (you should have one DocumentStore instance per application)
```go
func getDocumentStore(databaseName string) (*ravendb.DocumentStore, error) {
	serverNodes := []string{"http://live-test.ravendb.net"}
	store := ravendb.NewDocumentStore(serverNodes, databaseName)
	if err := store.Initialize(); err != nil {
		return nil, err
	}
	return store, nil
}
```
3. Open a session and close it when done
```go
session, err = store.OpenSession()
if err != nil {
	log.Fatalf("store.OpenSession() failed with %s", err)
}
// ... use session
session.Close()
```
4. Call `SaveChanges()` to persist changes in a session:
```go
var e *northwind.Employee
err = session.Load(&e, "employees/7-A")
if err != nil {
    log.Fatalf("session.Load() failed with %s\n", err)
}

origName := e.FirstName
e.FirstName = e.FirstName + "Changed"
err = session.Store(e)
if err != nil {
    log.Fatalf("session.Store() failed with %s\n", err)
}

err = session.SaveChanges()
if err != nil {
    log.Fatalf("session.SaveChanges() failed with %s\n", err)
}

var e2 *northwind.Employee
err = session.Load(&e2, "employees/7-A")
if err != nil {
    log.Fatalf("session.Load() failed with %s\n", err)
}
fmt.Printf("Updated Employee.FirstName from '%s' to '%s'\n", origName, e2.FirstName)
```
See `loadUpdateSave()` in [examples/main.go](examples/main.go) for full example.

## CRUD example

### Storing documents
```go
product := &northwind.Product{
    Name:         "iPhone X",
    PricePerUnit: 999.99,
    Category:     "electronis",
    ReorderLevel: 15,
}
err = session.Store(product)
if err != nil {
    log.Fatalf("session.Store() failed with %s\n", err)
}
```
See `crudStore()` in [examples/main.go](examples/main.go) for full example.


### Loading documents

```go
var e *northwind.Employee
err = session.Load(&e, "employees/7-A")
if err != nil {
    log.Fatalf("session.Load() failed with %s\n", err)
}
fmt.Printf("employee: %#v\n", e)
```
See `crudLoad()` in [examples/main.go](examples/main.go) for full example.

### Loading documents with includes

Some entities point to other entities via id. For example `Employee` has `ReportsTo` field which is an id of `Employee` that it reports to.

To improve performance by minimizing number of server requests, we can use includes functionality to load such linked entities.

```go
// load employee with id "employees/7-A" and entity whose id is ReportsTo
var e *northwind.Employee
err = session.Include("ReportsTo").Load(&e, "employees/5-A")
if err != nil {
    log.Fatalf("session.Load() failed with %s\n", err)
}
if e.ReportsTo == "" {
    fmt.Printf("Employee with id employees/5-A doesn't report to anyone\n")
    return
}

numRequests := session.GetNumberOfRequests()
var reportsTo *northwind.Employee
err = session.Load(&reportsTo, e.ReportsTo)
if err != nil {
    log.Fatalf("session.Load() failed with %s\n", err)
}
if numRequests != session.GetNumberOfRequests() {
    fmt.Printf("Something's wrong, this shouldn't send a request to the server\n")
} else {
    fmt.Printf("Loading e.ReportsTo employee didn't require a new request to the server because we've loaded it in original requests thanks to using Include functionality\n")
}
```
See `crudLoadWithInclude()` in [examples/main.go](examples/main.go) for full example.

### Updating documents

```go
// load entity from the server
var p *northwind.Product
err = session.Load(&p, productID)
if err != nil {
    log.Fatalf("session.Load() failed with %s\n", err)
}

// update price
origPrice = p.PricePerUnit
newPrice = origPrice + 10
p.PricePerUnit = newPrice
err = session.Store(p)
if err != nil {
    log.Fatalf("session.Store() failed with %s\n", err)
}

// persist changes on the server
err = session.SaveChanges()
if err != nil {
    log.Fatalf("session.SaveChanges() failed with %s\n", err)
}
```
See `crudUpdate()` in [examples/main.go](examples/main.go) for full example.

### Deleting documents

Deleting using entity

```go
// store a product and remember its id in productID

var p *northwind.Product
err = session.Load(&p, productID)
if err != nil {
    log.Fatalf("session.Load() failed with %s\n", err)
}

err = session.DeleteEntity(p)
if err != nil {
    log.Fatalf("session.Delete() failed with %s\n", err)
}

err = session.SaveChanges()
if err != nil {
    log.Fatalf("session.SaveChanges() failed with %s\n", err)
}

```

Deleting using id

```go
// store a product and remember its id in productID

err = session.Delete(productID)
if err != nil {
    log.Fatalf("session.Delete() failed with %s\n", err)
}

err = session.SaveChanges()
if err != nil {
    log.Fatalf("session.SaveChanges() failed with %s\n", err)
}
```
See `crudDeleteUsingID()` in [examples/main.go](examples/main.go) for full example.

