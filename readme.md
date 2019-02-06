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
4. Call `SaveChanges()` once you're done:
```go
session
 .load('users/1-A')
 .then((user) => {
   user.password = PBKDF2('new password');
 })
 .then(() => session.saveChanges())
 .then(() => {
    // data is persisted
    // you can proceed e.g. finish web request
  });
   