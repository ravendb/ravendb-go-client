package ravendb

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync/atomic"
	"testing"
	"time"
)

var (
	dbIndex             int32
	globalServer        *DocumentStore
	globalServerProcess *exec.Cmd

	globalSecuredServer *DocumentStore
)

func getDocumentStore() (*DocumentStore, error) {
	return getDocumentStoreWithName("test_db")
}

func getDocumentStoreWithName(dbName string) (*DocumentStore, error) {
	return getDocumentStore2(dbName, false, 0)

}

func getGlobalServer(secured bool) *DocumentStore {
	if secured {
		return globalSecuredServer
	}
	return globalServer
}

func runServer(secured bool) error {
	panicIf(true, "NYI")
	// TODO: implement me
	return nil
}

func getDocumentStore2(dbName string, secured bool, waitForIndexingTimeout time.Duration) (*DocumentStore, error) {
	n := atomic.AddInt32(&dbIndex, 1)
	name := fmt.Sprintf("%s_%d", dbName, n)
	documentStore := getGlobalServer(secured)
	if documentStore == nil {
		err := runServer(secured)
		if err != nil {
			return nil, err
		}
	}

	documentStore = getGlobalServer(secured)
	databaseRecord := NewDatabaseRecord()
	databaseRecord.DatabaseName = name
	// TODO: databaseRecord

	return nil, errors.New("NYI")
}

func shutdownTests() {
	// TODO: kill server process
}

func TestMain(m *testing.M) {
	code := m.Run()
	shutdownTests()
	os.Exit(code)
}
