package ravendb

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	//fmt.Printf("getDocumentStore2\n")
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
			fmt.Printf("runServer failed with %s\n", err)
			return nil, err
		}
	}

	documentStore = getGlobalServer(secured)
	databaseRecord := NewDatabaseRecord()
	databaseRecord.DatabaseName = name

	createDatabaseOperation := NewCreateDatabaseOperation(databaseRecord)
	_, err := documentStore.maintenance().server().send(createDatabaseOperation)
	if err != nil {
		return nil, err
	}

	urls := documentStore.getUrls()
	store := NewDocumentStoreWithUrlsAndDatabase(urls, name)

	if false && secured {
		// TODO: store.setCertificate(getTestClientCertificate());
	}

	// TODO: is over-written by CustomSerializationTest
	// customizeStore(store);
	// TODO:         hookLeakedConnectionCheck(store);

	// TODO:         setupDatabase(store);
	_, err = store.Initialize()
	if err != nil {
		return nil, err
	}

	// TODO:         store.addAfterCloseListener(((sender, event) -> {

	if waitForIndexingTimeout > 0 {
		waitForIndexing(store, name, waitForIndexingTimeout)
	}

	// TODO:    documentStores.add(store);

	return store, nil
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
		if RavenServerVerbose {
			fmt.Printf("%s\n", s)
		}
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

	if RavenServerVerbose {
		go func() {
			_, err := io.Copy(os.Stdout, proc.stdoutReader)
			if !(err == nil || err == io.EOF) {
				fmt.Printf("io.Copy() failed with %s\n", err)
			}
		}()
	}

	time.Sleep(time.Second) // TODO: probably not necessary

	store := NewDocumentStore()
	store.setUrls([]string{url})
	store.setDatabase("test.manager")
	store.getConventions().setDisableTopologyUpdates(true)

	if secured {
		panicIf(true, "NYI")
		globalSecuredServer = store
		//TODO: KeyStore clientCert = getTestClientCertificate();
		//TODO: store.setCertificate(clientCert);
	} else {
		globalServer = store
	}
	_, err = store.Initialize()
	return err
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
	noDb := os.Getenv("RAVEN_GO_NO_DB_TESTS")
	if noDb == "" {
		// this helps running tests from withing Visual Studio Code,
		// where env variables are not set
		serverPath := os.Getenv("RAVENDB_JAVA_TEST_SERVER_PATH")
		if serverPath == "" {
			home := os.Getenv("HOME")
			path := filepath.Join(home, "Documents", "RavenDB", "Server", "Raven.Server")
			_, err := os.Stat(path)
			if err != nil {
				cwd, err := os.Getwd()
				must(err)
				path = filepath.Join(cwd, "RavenDB", "Server", "Raven.Server")
				_, err = os.Stat(path)
				must(err)
			}
			os.Setenv("RAVENDB_JAVA_TEST_SERVER_PATH", path)
			fmt.Printf("Setting RAVENDB_JAVA_TEST_SERVER_PATH to '%s'\n", path)
		}
	}

	//RavenServerVerbose = true
	if useProxy() {
		go proxy.Run("")
	}
	code := m.Run()
	shutdownTests()

	if useProxy() {
		proxy.CloseLogFile()
		fmt.Printf("Closed proxy log file\n")
	}
	os.Exit(code)
}
