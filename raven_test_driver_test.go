package ravendb

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ravendb/ravendb-go-client/pkg/capture"
	"github.com/stretchr/testify/assert"
)

var (
	gRavenTestDriver *RavenTestDriver

	// in Java those are static fields of RavenTestDriver
	globalServer               *DocumentStore
	globalServerProcess        *Process
	globalSecuredServer        *DocumentStore
	globalSecuredServerProcess *Process
	index                      AtomicInteger
)

type RavenTestDriver struct {
	documentStores sync.Map // *DocumentStore => bool

	pcapPath   string
	pcapCloser io.Closer

	disposed bool
}

func NewRavenTestDriver() *RavenTestDriver {
	return &RavenTestDriver{}
}

func NewRavenTestDriverWithPacketCapture(pcapPath string) *RavenTestDriver {
	return &RavenTestDriver{pcapPath: pcapPath}
}

func (d *RavenTestDriver) getSecuredDocumentStore() (*DocumentStore, error) {
	return d.getDocumentStore2("test_db", true, 0)
}

// func (d *RavenTestDriver)
func (d *RavenTestDriver) getTestClientCertificate() *KeyStore {
	// TODO: implement me
	return nil
}

func (d *RavenTestDriver) getDocumentStore() (*DocumentStore, error) {
	return d.getDocumentStoreWithName("test_db")
}

func (d *RavenTestDriver) getSecuredDocumentStoreWithName(database string) (*DocumentStore, error) {
	return d.getDocumentStore2(database, true, 0)
}

func (d *RavenTestDriver) getDocumentStoreWithName(dbName string) (*DocumentStore, error) {
	return d.getDocumentStore2(dbName, false, 0)
}

func (d *RavenTestDriver) getDocumentStore2(dbName string, secured bool, waitForIndexingTimeout time.Duration) (*DocumentStore, error) {
	//fmt.Printf("getDocumentStore2\n")

	n := index.incrementAndGet()
	name := fmt.Sprintf("%s_%d", dbName, n)
	documentStore := d.getGlobalServer(secured)
	if documentStore == nil {
		err := d.runServer(secured)
		if err != nil {
			fmt.Printf("runServer failed with %s\n", err)
			return nil, err
		}
	}

	documentStore = d.getGlobalServer(secured)
	databaseRecord := NewDatabaseRecord()
	databaseRecord.DatabaseName = name

	createDatabaseOperation := NewCreateDatabaseOperation(databaseRecord)
	err := documentStore.maintenance().server().send(createDatabaseOperation)
	if err != nil {
		return nil, err
	}

	urls := documentStore.getUrls()
	store := NewDocumentStoreWithUrlsAndDatabase(urls, name)

	if secured {
		store.setCertificate(d.getTestClientCertificate())
	}

	// TODO: is over-written by CustomSerializationTest
	// customizeStore(store);
	d.hookLeakedConnectionCheck(store)

	d.setupDatabase(store)
	_, err = store.Initialize()
	if err != nil {
		return nil, err
	}

	fn := func(store *DocumentStore) {
		_, ok := d.documentStores.Load(store)
		if !ok {
			// TODO: shouldn't happen?
			return
		}

		operation := NewDeleteDatabasesOperation(store.getDatabase(), true)
		store.maintenance().server().send(operation)
	}

	store.addAfterCloseListener(fn)

	if waitForIndexingTimeout > 0 {
		d.waitForIndexing(store, name, waitForIndexingTimeout)
	}

	d.documentStores.Store(store, true)

	return store, nil
}

func (d *RavenTestDriver) hookLeakedConnectionCheck(store *DocumentStore) {
	// TODO: no-op for now. Not sure if I have enough info
	// to replicate this functionality in Go
}

// Note: it's virtual in Java but there's only one implementation
// that is a no-op
func (d *RavenTestDriver) setupDatabase(documentStore *DocumentStore) {
	// empty by design
}

func (d *RavenTestDriver) runServer(secured bool) error {
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
	d.setGlobalServerProcess(secured, proc)

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

	// capture packets if not https
	if !secured && d.pcapPath != "" {
		ipAddr := strings.TrimPrefix(url, "http://")
		fmt.Printf("Capturing packets from interface '%s' to file '%s'\n", ipAddr, d.pcapPath)
		d.pcapCloser, err = capture.StartCapture(ipAddr, d.pcapPath)
		if err != nil {
			if strings.Contains(err.Error(), "You don't have permission") {
				// ignore if this is a permissions error
				fmt.Printf("Failed to start packet capture, error: '%s'\n", err)
				fmt.Printf("To get capture, re-run under root e.g. with:\n")
				fmt.Printf("sudo -E ./run_single_test.sh\n")
			} else {
				panic(err)
			}
		}
	}

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
		panic("NYI")
		globalSecuredServer = store
		//TODO: KeyStore clientCert = getTestClientCertificate();
		//TODO: store.setCertificate(clientCert);
	} else {
		globalServer = store
	}
	_, err = store.Initialize()
	return err
}

func (d *RavenTestDriver) waitForIndexing(store *DocumentStore, database string, timeout time.Duration) error {
	admin := store.maintenance().forDatabase(database)
	if timeout == 0 {
		timeout = time.Minute
	}

	sp := time.Now()
	for time.Since(sp) < timeout {
		op := NewGetStatisticsOperation()
		err := admin.send(op)
		if err != nil {
			return err
		}
		databaseStatistics := op.Command.Result
		isDone := true
		hasError := false
		for _, index := range databaseStatistics.Indexes {
			if index.getState() == IndexState_DISABLED {
				continue
			}
			if index.isStale() || strings.HasPrefix(index.getName(), Constants_Documents_Indexing_SIDE_BY_SIDE_INDEX_NAME_PREFIX) {
				isDone = false
			}
			if index.getState() == IndexState_ERROR {
				hasError = true
			}
		}
		if isDone {
			return nil
		}
		if hasError {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}

	op := NewGetIndexErrorsOperation(nil)
	err := admin.send(op)
	if err != nil {
		return err
	}
	allIndexErrorsText := ""
	/*
		// TODO: port this
		Function<IndexErrors, String> formatIndexErrors = indexErrors -> {
				String errorsListText = Arrays.stream(indexErrors.getErrors()).map(x -> "-" + x).collect(Collectors.joining(System.lineSeparator()));
				return "Index " + indexErrors.getName() + " (" + indexErrors.getErrors().length + " errors): "+ System.lineSeparator() + errorsListText;
			};

			if (errors != null && errors.length > 0) {
				allIndexErrorsText = Arrays.stream(errors).map(x -> formatIndexErrors.apply(x)).collect(Collectors.joining(System.lineSeparator()));
			}
	*/
	return NewTimeoutException("The indexes stayed stale for more than %s.%s", timeout, allIndexErrorsText)
}

func (d *RavenTestDriver) killGlobalServerProcess(secured bool) {
	var err error
	if secured {
		if globalSecuredServerProcess != nil {
			err = globalSecuredServerProcess.cmd.Process.Kill()
			if err != nil {
				fmt.Printf(" -- globalSecuredServerProcess.cmd.Process.Kill() failed with '%s'\n", err)
			}
			globalSecuredServerProcess = nil
		}
	} else {
		if globalServerProcess != nil {
			globalServerProcess.cmd.Process.Kill()
			if err != nil {
				fmt.Printf(" -- globalServerProcess.cmd.Process.Kill() failed with '%s'\n", err)
			}
			globalServerProcess = nil
		}
	}
}

func (d *RavenTestDriver) getGlobalServer(secured bool) *DocumentStore {
	if secured {
		return globalSecuredServer
	}
	return globalServer
}

func (d *RavenTestDriver) setGlobalServerProcess(secured bool, p *Process) {
	if secured {
		globalSecuredServerProcess = p
	} else {
		globalServerProcess = p
	}
}

func (d *RavenTestDriver) Close() {
	if d.disposed {
		return
	}

	fn := func(key, value interface{}) bool {
		documentStore := key.(*DocumentStore)
		documentStore.Close()
		return true
	}
	d.documentStores.Range(fn)
	d.disposed = true
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

func shutdownTests() {
	gRavenTestDriver.killGlobalServerProcess(true)
	gRavenTestDriver.killGlobalServerProcess(false)
}

var dbTestsDisabledAlreadyPrinted = false

func dbTestsDisabled() bool {
	if os.Getenv("RAVEN_GO_NO_DB_TESTS") != "" {
		if !dbTestsDisabledAlreadyPrinted {
			dbTestsDisabledAlreadyPrinted = true
			fmt.Printf("DB tests are disabled\n")
		}
		return true
	}
	return false
}

func getDocumentStoreMust(t *testing.T) *DocumentStore {
	store, err := gRavenTestDriver.getDocumentStore()
	assert.NoError(t, err)
	assert.NotNil(t, store)
	return store
}

func openSessionMust(t *testing.T, store *DocumentStore) *DocumentSession {
	session, err := store.OpenSession()
	assert.NoError(t, err)
	assert.NotNil(t, session)
	return session
}

// In Java, RavenTestDriver is created/destroyed for each test
// In Go we have to do it manually

func createTestDriver() {
	panicIf(gRavenTestDriver != nil, "gravenTestDriver must be nil")
	gRavenTestDriver = NewRavenTestDriver()
}

func createTestDriverWithPacketCapture(pcapPath string) {
	panicIf(gRavenTestDriver != nil, "gravenTestDriver must be nil")
	gRavenTestDriver = NewRavenTestDriverWithPacketCapture(pcapPath)
}

func deleteTestDriver() {
	if gRavenTestDriver == nil {
		return
	}
	gRavenTestDriver.Close()
	if gRavenTestDriver.pcapCloser != nil {
		fmt.Printf("Closing pcap capture\n")
		gRavenTestDriver.pcapCloser.Close()
		gRavenTestDriver.pcapCloser = nil
		fmt.Printf("Closed pcap capture\n")
	}
	gRavenTestDriver.killGlobalServerProcess(true)
	gRavenTestDriver.killGlobalServerProcess(false)
	gRavenTestDriver = nil
}

// This helps debugging leaking gorutines by dumping stack traces
// of all goroutines to a file
func logGoroutines(file string) {
	if file == "" {
		file = "goroutines.txt"
	}
	path := filepath.Join("logs", file)
	dir := filepath.Dir(path)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return
	}
	profile := pprof.Lookup("goroutine")
	if profile == nil {
		return
	}

	f, err := os.Create(path)
	if err != nil {
		return
	}
	defer f.Close()
	profile.WriteTo(f, 2)
}

func TestMain(m *testing.M) {
	if os.Getenv("VERBOSE_LOG") != "" {
		verboseLog = true
	}

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

	var code int

	// make sure it's called even if panic happens
	defer func() {
		shutdownTests()

		//logGoroutines()
		os.Exit(code)
	}()

	code = m.Run()
}
