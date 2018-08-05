package ravendb

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"sync"
	"syscall"
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

	pcapPath            string
	pcapCapturerProcess *exec.Cmd

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

func startPcapCaptureProcess(ipAddr string, pcapPath string) (*exec.Cmd, error) {
	cmd := exec.Command("./capturer", "-addr", ipAddr, "-pcap", pcapPath, "-show-request-summary")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func stopPcapCaptureProcess(cmd *exec.Cmd) {
	if cmd == nil {
		return
	}
	// capture process should exit cleanly when given SIGINT
	// we wait up to 3 seconds for exit
	cmd.Process.Signal(syscall.SIGINT)
	exited := make(chan bool, 1)
	go func() {
		cmd.Wait()
		exited <- true
	}()

	select {
	case <-exited:
		fmt.Printf("Pcap capture process exited cleanly\n")
	case <-time.After(time.Second * 3):
		fmt.Printf("Pcap capture process didn't exit within 3 seconds\n")
		cmd.Process.Kill()
	}
}

func (d *RavenTestDriver) killCaptureProcess() {
	if d == nil {
		return
	}
	stopPcapCaptureProcess(d.pcapCapturerProcess)
	d.pcapCapturerProcess = nil
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
	proc, err := RavenServerRunner_run(locator)
	if err != nil {
		fmt.Printf("RavenServerRunner_run failed with %s\n", err)
		return err
	} else {
		args := strings.Join(proc.cmd.Args, " ")
		fmt.Printf("Started raven server '%s'\n", args)
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
		capture.RequestsSummaryWriter = os.Stdout
		d.pcapCapturerProcess, err = startPcapCaptureProcess(ipAddr, d.pcapPath)
		if err != nil {
			fmt.Printf("Failed to start pcap capturer process. Error: %s\n", err)
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

func killServer(procPtr **Process) {
	proc := *procPtr
	if proc == nil {
		return
	}
	err := proc.cmd.Process.Kill()
	if err != nil {
		fmt.Printf("cmd.Process.Kill() failed with '%s'\n", err)
	} else {
		s := strings.Join(proc.cmd.Args, " ")
		fmt.Printf("Killed a process '%s'\n", s)
	}
	*procPtr = nil
}

func killGlobalServerProcesses() {
	killServer(&globalSecuredServerProcess)
	killServer(&globalServerProcess)
	globalSecuredServer = nil
	globalServer = nil
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

func shutdownTests() {
	killGlobalServerProcesses()
	gRavenTestDriver.killCaptureProcess()
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

func isUpper(c byte) bool {
	return c >= 'A' && c <= 'Z'
}

// converts "TestIndexesFromClient" => "indexes_from_client"
func testNameToFileName(s string) string {
	s = strings.TrimPrefix(s, "Test")
	lower := strings.ToLower(s)
	var res []byte
	n := len(s)
	for i := 0; i < n; i++ {
		c := s[i]
		if i > 0 && isUpper(c) {
			res = append(res, '_')
		}
		res = append(res, lower[i])
	}
	return string(res)
}

func pcapPathFromTestName(t *testing.T) string {
	name := "trace_" + testNameToFileName(t.Name()) + "_go.pcap"
	path := filepath.Join("logs", name)
	os.Mkdir("logs", 0755)
	return path
}

func ravenLogsDirFromTestName(t *testing.T) string {
	// if this is not full path, raven will put it in it's own Logs directory
	// next to server executable
	cwd, _ := os.Getwd()
	name := "trace_" + testNameToFileName(t.Name()) + "_go_server_dir"
	path := filepath.Join(cwd, "logs", name)
	// recreate dir for clean logs
	os.RemoveAll(path)
	os.MkdirAll(path, 0755)
	return path
}

func deleteTestDriver() {
	if gRavenTestDriver == nil {
		return
	}
	gRavenTestDriver.Close()
	killGlobalServerProcesses()
	gRavenTestDriver.killCaptureProcess()
	gRavenTestDriver = nil
}

func maybeConvertPcapToTxt(pcapPath string) {
	if pcapPath == "" {
		return
	}
	exe := "./pcap_convert"
	_, err := os.Stat(exe)
	if err != nil {
		// skip if we don't have pcap_convert
		return
	}
	cmd := exec.Command(exe, pcapPath)
	err = cmd.Run()
	if err != nil {
		s := strings.Join(cmd.Args, " ")
		fmt.Printf("command '%s' failed with '%s'\n", s, err)
	}
}

func createTestDriver(t *testing.T) func() {
	panicIf(gRavenTestDriver != nil, "gravenTestDriver must be nil")

	maybeEnableVerbose()
	gRavenLogsDir = ravenLogsDirFromTestName(t)

	fmt.Printf("\nStarting %s\n", t.Name())
	var pcapPath string
	if os.Getenv("PCAP_CAPTURE") != "" {
		pcapPath = pcapPathFromTestName(t)
		panicIf(gRavenTestDriver != nil, "gravenTestDriver must be nil")
		gRavenTestDriver = NewRavenTestDriverWithPacketCapture(pcapPath)
	} else {
		gRavenTestDriver = NewRavenTestDriver()
	}
	return func() {
		deleteTestDriver()
		maybeConvertPcapToTxt(pcapPath)
	}
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

func maybeEnableVerbose() {
	if os.Getenv("VERBOSE_LOG") != "" {
		verboseLog = true
		fmt.Printf("verbose logging enabled\n")
	}
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

	var code int

	// make sure it's called even if panic happens
	defer func() {
		shutdownTests()

		//logGoroutines()
		os.Exit(code)
	}()

	code = m.Run()
}
