package ravendb

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
)

var (
	dbIndex                    int32
	globalServer               *DocumentStore
	globalServerProcess        *Process
	globalSecuredServerProcess *Process

	globalSecuredServer *DocumentStore
)

func getGlobalServer(secured bool) *DocumentStore {
	if secured {
		return globalSecuredServer
	}
	return globalServer
}

func setGlobalServerProcess(secured bool, p *Process) {
	if secured {
		globalSecuredServerProcess = p
	} else {
		globalServerProcess = p
	}
}

func getDocumentStore() (*DocumentStore, error) {
	return getDocumentStoreWithName("test_db")
}

func getDocumentStoreWithName(dbName string) (*DocumentStore, error) {
	return getDocumentStore2(dbName, false, 0)

}

func getDocumentStore2(dbName string, secured bool, waitForIndexingTimeout time.Duration) (*DocumentStore, error) {
	fmt.Printf("getDocumentStore2\n")
	// when db tests are disabled we return nil DocumentStore which is a signal
	// to the caller to skip the db tests
	if os.Getenv("RAVEN_GO_NO_DB_TESTS") != "" {
		fmt.Printf("DB tests are disabled\n")
		return nil, nil
	}

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

	createDatabaseOperation := NewCreateDatabaseOperation(databaseRecord)
	exec := documentStore.maintenance().requestExecutor.GetCommandExecutor(false)
	_, err := ExecuteCreateDatabaseCommand(exec, createDatabaseOperation)
	if err != nil {
		return nil, err
	}

	urls := documentStore.getURLS()
	store := NewDocumentStore(urls, name)

	if false && secured {
		// TODO: store.setCertificate(getTestClientCertificate());
	}

	// TODO: is over-written by CustomSerializationTest
	// customizeStore(store);
	// TODO:         hookLeakedConnectionCheck(store);

	// TODO:         setupDatabase(store);
	err = store.Initialize()
	if err != nil {
		return nil, err
	}

	// TODO:         store.addAfterCloseListener(((sender, event) -> {

	if waitForIndexingTimeout > 0 {
		waitForIndexing(store, name, waitForIndexingTimeout)
	}

	// TODO:    documentStores.add(store);

	return store, errors.New("NYI")
}

func waitForIndexing(store *DocumentStore, database String, timeout time.Duration) {
	// TODO: implement me
	panicIf(true, "NYI")
}

func runServer(secured bool) error {
	var locator *RavenServerLocator
	var err error
	if secured {
		locator, err = NewSecuredServiceLocator()
	} else {
		locator, err = NewTestServiceLocator()
	}
	if err != nil {
		return err
	}
	fmt.Printf("runServer: starting server\n")
	proc, err := RavenServerRunner_run(locator)
	if err != nil {
		fmt.Printf("RavenServerRunner_run failed with %s\n", err)
		return err
	}
	setGlobalServerProcess(secured, proc)

	// parse stdout of the server to extract server listening port from line:
	// Server available on: http://127.0.0.1:50386
	wantedPrefix := "Server available on: "
	scanner := bufio.NewScanner(proc.stdoutReader)
	timeStart := time.Now()
	url := ""
	for scanner.Scan() {
		dur := time.Since(timeStart)
		if dur > time.Minute {
			break
		}
		s := scanner.Text()
		// fmt.Printf("line: '%s'\n", s)
		if !strings.HasPrefix(s, wantedPrefix) {
			continue
		}
		s = strings.TrimPrefix(s, wantedPrefix)
		url = strings.TrimSpace(s)
		break
	}
	if scanner.Err() != nil {
		return scanner.Err()
	}
	if url == "" {
		return fmt.Errorf("Unable to start server")
	}
	fmt.Printf("Server started on: '%s'\n", url)

	urls := []string{url}
	store := NewDocumentStore(urls, "test.manager")
	//store.setUrls([]string{url})
	//store.setDatabase("test.manager")
	store.getConventions().setDisableTopologyUpdates(true)

	if secured {
		panicIf(true, "NYI")
		globalSecuredServer = store
		//TODO: KeyStore clientCert = getTestClientCertificate();
		//TODO: store.setCertificate(clientCert);
	} else {
		globalServer = store
	}
	return store.Initialize()
}

func killGlobalServerProcess(secured bool) {
	if secured {
		if globalSecuredServerProcess != nil {
			globalSecuredServerProcess.cmd.Process.Kill()
			globalSecuredServerProcess = nil
		}
	} else {
		if globalServerProcess != nil {
			globalServerProcess.cmd.Process.Kill()
			globalServerProcess = nil
		}
	}
}

func shutdownTests() {
	killGlobalServerProcess(true)
	killGlobalServerProcess(false)
}

var (
	useProxyCached *bool
)

func useProxy() bool {
	if useProxyCached != nil {
		return *useProxyCached
	}
	var b bool
	if os.Getenv("HTTP_PROXY") != "" {
		fmt.Printf("Using http proxy\n")
		b = true
	} else {
		fmt.Printf("Not using http proxy\n")
	}
	useProxyCached = &b
	return b
}

func TestMain(m *testing.M) {
	fmt.Printf("TestMain\n")
	if useProxy() {
		logFileTmpl := "trace_0_start_go.txt"
		go proxy.Run(logFileTmpl)
	}
	code := m.Run()
	shutdownTests()

	if useProxy() {
		proxy.CloseLogFile()
		fmt.Printf("Closed proxy log file\n")
	}
	os.Exit(code)
}
