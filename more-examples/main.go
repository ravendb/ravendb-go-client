package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/ravendb/ravendb-go-client"
	"github.com/ravendb/ravendb-go-client/examples/northwind"
	"log"
	"time"
)

// "Demo" is a Northwind sample database
// You can browse its content via web interface at
// http://live-test.ravendb.net/studio/index.html#databases/documents?&database=Demo
var (
	dbName        = "TestDapr"
	serverNodeURL = "http://127.0.0.1:8080"

	secureNodeURL = "https://a.free.nemanja.ravendb.cloud/"

	// if true, we'll show summary of HTTP requests made to the server
	// and dump full info about failed HTTP requests
	verboseLogging = true

	// if true, logs all http requests/responses to a file for further inspection
	// this is for use in tests so the file has a fixed location:
	// logs/trace_${test_name}_go.txt
	logAllRequests = false

	// if logAllRequests is true, this is a path of a file where we log
	// info about all HTTP requests
	logAllRequestsPath = "http_requests_log.txt"
)

func getDocumentStore(databaseName string) (*ravendb.DocumentStore, error) {
	serverNodes := []string{serverNodeURL}
	store := ravendb.NewDocumentStore(serverNodes, databaseName)
	if err := store.Initialize(); err != nil {
		return nil, err
	}
	return store, nil
}

func openSession(databaseName string) (*ravendb.DocumentStore, *ravendb.DocumentSession, error) {
	store, err := getDocumentStore(dbName)
	if err != nil {
		return nil, nil, fmt.Errorf("getDocumentStore() failed with %s", err)
	}

	session, err := store.OpenSession("")
	if err != nil {
		return nil, nil, fmt.Errorf("store.OpenSession() failed with %s", err)
	}
	return store, session, nil
}

func getSecureDocumentStore(databaseName string) (*ravendb.DocumentStore, error) {
	cerPath := "/Users/nemanjamalocic/projects/RavenDB/ravendb-go-client/more-examples/PEM/free.nemanja.client.certificate.crt"
	keyPath := "/Users/nemanjamalocic/projects/RavenDB/ravendb-go-client/more-examples/PEM/free.nemanja.client.certificate.key"
	cer, err := tls.LoadX509KeyPair(cerPath, keyPath)
	if err != nil {
		return nil, err
	}
	serverNodes := []string{secureNodeURL}
	store := ravendb.NewDocumentStore(serverNodes, databaseName)
	store.Certificate = &cer
	x509cert, err := x509.ParseCertificate(cer.Certificate[0])
	if err != nil {
		return nil, err
	}
	store.TrustStore = x509cert
	if store.TrustStore == nil {
		panic("nil trust store")
	}
	if err := store.Initialize(); err != nil {
		return nil, err
	}
	return store, nil
}

func openSecureSession(databaseName string) (*ravendb.DocumentStore, *ravendb.DocumentSession, error) {
	store, err := getSecureDocumentStore(dbName)
	if err != nil {
		return nil, nil, fmt.Errorf("getDocumentStore() failed with %s", err)
	}

	session, err := store.OpenSession("")
	if err != nil {
		return nil, nil, fmt.Errorf("store.OpenSession() failed with %s", err)
	}
	return store, session, nil
}

func addNewDocumentUsingSecureConnection() {
	store, session, err := openSecureSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	item1 := &northwind.Employee{
		ID:        "1",
		FirstName: "Test name",
	}
	err = session.Store(item1)
	item2 := &northwind.Employee{
		ID:        "2",
		FirstName: "Nemanja",
	}
	err = session.Store(item2)
	if err != nil {
		log.Fatalf("session.Store() failed with %s\n", err)
	}
	err = session.SaveChanges()
	if err != nil {
		log.Fatalf("Failed to get save changes")
	}
	var results = make(map[string]*northwind.Employee, 2)
	var keys = make([]string, 2)
	keys[0] = "1"
	keys[1] = "2"
	err = session.LoadMulti(results, keys)
	if err != nil {
		log.Fatalf("Failed to get save changes")
	}
}

func addNewDocumentWithChangeVector() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	item := &northwind.Employee{
		FirstName: "Test name",
	}

	err = session.StoreWithID(item, "1")
	err = session.SaveChanges()
	if err != nil {
		log.Fatalf("Failed to get save changes")
	}

	metadata, err := session.Advanced().GetMetadataFor(item)
	cVector, success := metadata.Get("@change-vector")

	if !success {
		log.Fatalf("error getting cVector", cVector)
	}

	item2 := &northwind.Employee{
		FirstName: "Updated name",
	}

	err = session.StoreWithChangeVectorAndID(item2, cVector.(string), "1")
	if err != nil {
		log.Fatalf("session.Store() failed with %s\n", err)
	}

	err = session.SaveChanges()
	if err != nil {
		log.Fatalf("Failed to get save changes")
	}
}

func addNewDocumentWithTimeToLive() {
	store, session, err := openSession(dbName)
	if err != nil {
		log.Fatalf("openSession() failed with %s\n", err)
	}
	defer store.Close()
	defer session.Close()

	item := &northwind.Employee{
		FirstName: "Test name",
	}
	err = session.Store(item)
	if err != nil {
		log.Fatalf("session.Store() failed with %s\n", err)
	}
	metaData, err := session.Advanced().GetMetadataFor(item)
	if err != nil {
		log.Fatalf("Failed to get metadata for item")
	}
	expiry := time.Now().Add(time.Minute).UTC()
	// Format the time to ISO 8601 string
	iso8601String := expiry.Format("2006-01-02T15:04:05.9999999Z07:00")
	metaData.Put("@expires", iso8601String)

	err = session.SaveChanges()
	if err != nil {
		log.Fatalf("Failed to get save changes")
	}
}

func main() {
	//addNewDocumentWithTimeToLive()
	addNewDocumentWithChangeVector()
	// addNewDocumentUsingSecureConnection()
}
