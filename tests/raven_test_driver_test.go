package tests

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"runtime/pprof"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/ravendb/ravendb-go-client/pkg/capture"
	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

var (
	gRavenTestDriver *RavenTestDriver

	// in Java those are static fields of RavenTestDriver
	globalServer               *ravendb.DocumentStore
	globalServerProcess        *Process
	globalSecuredServer        *ravendb.DocumentStore
	globalSecuredServerProcess *Process
	index                      ravendb.AtomicInteger
)

func must(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func panicIf(cond bool, format string, args ...interface{}) {
	if cond {
		err := fmt.Errorf(format, args...)
		must(err)
	}
}

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

func (d *RavenTestDriver) getSecuredDocumentStore() (*ravendb.DocumentStore, error) {
	return d.getDocumentStore2("test_db", true, 0)
}

// func (d *RavenTestDriver)
func (d *RavenTestDriver) getTestClientCertificate() *ravendb.KeyStore {
	// TODO: implement me
	return nil
}

func (d *RavenTestDriver) getDocumentStore() (*ravendb.DocumentStore, error) {
	return d.getDocumentStoreWithName("test_db")
}

func (d *RavenTestDriver) getSecuredDocumentStoreWithName(database string) (*ravendb.DocumentStore, error) {
	return d.getDocumentStore2(database, true, 0)
}

func (d *RavenTestDriver) getDocumentStoreWithName(dbName string) (*ravendb.DocumentStore, error) {
	return d.getDocumentStore2(dbName, false, 0)
}

func (d *RavenTestDriver) getDocumentStore2(dbName string, secured bool, waitForIndexingTimeout time.Duration) (*ravendb.DocumentStore, error) {
	//fmt.Printf("getDocumentStore2\n")

	n := index.IncrementAndGet()
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
	databaseRecord := ravendb.NewDatabaseRecord()
	databaseRecord.DatabaseName = name

	createDatabaseOperation := ravendb.NewCreateDatabaseOperation(databaseRecord)
	err := documentStore.Maintenance().Server().Send(createDatabaseOperation)
	if err != nil {
		return nil, err
	}

	urls := documentStore.GetUrls()
	store := ravendb.NewDocumentStoreWithUrlsAndDatabase(urls, name)

	if secured {
		store.SetCertificate(d.getTestClientCertificate())
	}

	// TODO: is over-written by CustomSerializationTest
	// customizeStore(Store);
	d.hookLeakedConnectionCheck(store)

	d.setupDatabase(store)
	_, err = store.Initialize()
	if err != nil {
		return nil, err
	}

	fn := func(store *ravendb.DocumentStore) {
		_, ok := d.documentStores.Load(store)
		if !ok {
			// TODO: shouldn't happen?
			return
		}

		operation := ravendb.NewDeleteDatabasesOperation(store.GetDatabase(), true)
		store.Maintenance().Server().Send(operation)
	}

	store.AddAfterCloseListener(fn)

	if waitForIndexingTimeout > 0 {
		d.waitForIndexing(store, name, waitForIndexingTimeout)
	}

	d.documentStores.Store(store, true)

	return store, nil
}

func (d *RavenTestDriver) hookLeakedConnectionCheck(store *ravendb.DocumentStore) {
	// TODO: no-op for now. Not sure if I have enough info
	// to replicate this functionality in Go
}

// Note: it's virtual in Java but there's only one implementation
// that is a no-op
func (d *RavenTestDriver) setupDatabase(documentStore *ravendb.DocumentStore) {
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
		if ravendb.RavenServerVerbose {
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

	if ravendb.RavenServerVerbose {
		go func() {
			_, err := io.Copy(os.Stdout, proc.stdoutReader)
			if !(err == nil || err == io.EOF) {
				fmt.Printf("io.Copy() failed with %s\n", err)
			}
		}()
	}

	time.Sleep(time.Second) // TODO: probably not necessary

	store := ravendb.NewDocumentStore()
	store.SetUrls([]string{url})
	store.SetDatabase("test.manager")
	store.GetConventions().SetDisableTopologyUpdates(true)

	if secured {
		panic("NYI")
		globalSecuredServer = store
		//TODO: KeyStore clientCert = getTestClientCertificate();
		//TODO: Store.setCertificate(clientCert);
	} else {
		globalServer = store
	}
	_, err = store.Initialize()
	return err
}

func (d *RavenTestDriver) waitForIndexing(store *ravendb.DocumentStore, database string, timeout time.Duration) error {
	admin := store.Maintenance().ForDatabase(database)
	if timeout == 0 {
		timeout = time.Minute
	}

	sp := time.Now()
	for time.Since(sp) < timeout {
		op := ravendb.NewGetStatisticsOperation()
		err := admin.Send(op)
		if err != nil {
			return err
		}
		databaseStatistics := op.Command.Result
		isDone := true
		hasError := false
		for _, index := range databaseStatistics.Indexes {
			if index.GetState() == ravendb.IndexState_DISABLED {
				continue
			}
			if index.IsStale() || strings.HasPrefix(index.GetName(), ravendb.Constants_Documents_Indexing_SIDE_BY_SIDE_INDEX_NAME_PREFIX) {
				isDone = false
			}
			if index.GetState() == ravendb.IndexState_ERROR {
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

	op := ravendb.NewGetIndexErrorsOperation(nil)
	err := admin.Send(op)
	if err != nil {
		return err
	}
	allIndexErrorsText := ""
	/*
		// TODO: port this
		Function<IndexErrors, String> formatIndexErrors = indexErrors -> {
				String errorsListText = Arrays.stream(indexErrors.getErrors()).map(x -> "-" + x).collect(Collectors.joining(System.lineSeparator()));
				return "Index " + indexErrors.GetName() + " (" + indexErrors.getErrors().length + " errors): "+ System.lineSeparator() + errorsListText;
			};

			if (errors != null && errors.length > 0) {
				allIndexErrorsText = Arrays.stream(errors).map(x -> formatIndexErrors.apply(x)).collect(Collectors.joining(System.lineSeparator()));
			}
	*/
	return ravendb.NewTimeoutException("The indexes stayed stale for more than %s.%s", timeout, allIndexErrorsText)
}

func killServer(procPtr **Process) {
	proc := *procPtr
	if proc == nil {
		return
	}
	if proc.cmd.ProcessState != nil && proc.cmd.ProcessState.Exited() {
		fmt.Printf("RavenDB process has already exited with '%s'\n", proc.cmd.ProcessState)
	}
	err := proc.cmd.Process.Kill()
	if err != nil {
		fmt.Printf("cmd.Process.Kill() failed with '%s'\n", err)
	} else {
		s := strings.Join(proc.cmd.Args, " ")
		fmt.Printf("Killed RavenDB process %d '%s'\n", proc.cmd.Process.Pid, s)
	}
	*procPtr = nil
}

func killGlobalServerProcesses() {
	killServer(&globalSecuredServerProcess)
	killServer(&globalServerProcess)
	globalSecuredServer = nil
	globalServer = nil
}

func (d *RavenTestDriver) getGlobalServer(secured bool) *ravendb.DocumentStore {
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
		documentStore := key.(*ravendb.DocumentStore)
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

func getDocumentStoreMust(t *testing.T) *ravendb.DocumentStore {
	store, err := gRavenTestDriver.getDocumentStore()
	assert.NoError(t, err)
	assert.NotNil(t, store)
	return store
}

func openSessionMust(t *testing.T, store *ravendb.DocumentStore) *ravendb.DocumentSession {
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

func httpLogPathFromTestName(t *testing.T) string {
	name := "trace_" + testNameToFileName(t.Name()) + "_go.txt"
	path := filepath.Join("logs", name)
	os.Mkdir("logs", 0755)
	return path
}

func ravenLogsDirFromTestName(t *testing.T) string {
	// if this is not full path, raven will put it in it's own Logs directory
	// next to server executable
	cwd, _ := os.Getwd()
	name := testNameToFileName(t.Name()) + ".log.txt"
	path := filepath.Join(cwd, "logs", "server", "go", name)
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

var (
	ravendbWindowsDownloadURL = "https://daily-builds.s3.amazonaws.com/RavenDB-4.0.6-windows-x64.zip"
	ravenWindowsZipPath       = "Ravendb-4.0.6.zip"
)

func getRavendbExePath() string {
	cwd, err := os.Getwd()
	must(err)

	path := filepath.Join(cwd, "..", "RavenDB", "Server", "Raven.Server")
	if ravendb.IsWindows() {
		path += ".exe"
	}
	if ravendb.FileExists(path) {
		return path
	}

	path = filepath.Join(cwd, "RavenDB", "Server", "Raven.Server")
	if ravendb.IsWindows() {
		path += ".exe"
	}
	return path
}

func downloadServerIfNeededWindows() {
	_, err := os.Stat(getRavendbExePath())
	if err == nil {
		fmt.Printf("Server already present in %s\n", getRavendbExePath())
		return
	}
	_, err = os.Stat(ravenWindowsZipPath)
	if err != nil {
		fmt.Printf("Downloading %s...", ravendbWindowsDownloadURL)
		timeStart := time.Now()
		err := ravendb.HttpDl(ravendbWindowsDownloadURL, ravenWindowsZipPath)
		must(err)
		fmt.Printf(" took %s\n", time.Since(timeStart))
	}
	destDir := "RavenDB"
	fmt.Printf("Unzipping %s to %s...", ravenWindowsZipPath, destDir)
	timeStart := time.Now()
	err = ravendb.Unzip(ravenWindowsZipPath, destDir)
	must(err)
	fmt.Printf(" took %s\n", time.Since(timeStart))
}

func downloadServerIfNeeded() {
	if ravendb.IsWindows() {
		downloadServerIfNeededWindows()
		return
	}
}

// this helps running tests from withing Visual Studio Code,
// where env variables are not set
func detectServerPath() {
	// explicitly setting RAVEN_GO_NO_DB_TESTS=true disables database tests
	// so no need for the server
	noDB := os.Getenv("RAVEN_GO_NO_DB_TESTS")
	if noDB != "" {
		return
	}

	// use RAVENDB_JAVA_TEST_SERVER_PATH env var if explicilty set
	serverPath := os.Getenv("RAVENDB_JAVA_TEST_SERVER_PATH")
	if serverPath != "" {
		return
	}

	path := getRavendbExePath()
	_, err := os.Stat(path)
	must(err)
	os.Setenv("RAVENDB_JAVA_TEST_SERVER_PATH", path)
	fmt.Printf("Setting RAVENDB_JAVA_TEST_SERVER_PATH to '%s'\n", path)
}

func maybePrintFailedRequestsLog() {
	if ravendb.LogFailedRequests && ravendb.LogFailedRequestsDelayed {
		buf := ravendb.HTTPFailedRequestsLogger.(*bytes.Buffer)
		os.Stdout.Write(buf.Bytes())
		buf.Reset()
	}
}

// returns a shutdown function that must be called to cleanly shutdown test
func createTestDriver(t *testing.T) func() {
	panicIf(gRavenTestDriver != nil, "gravenTestDriver must be nil")
	downloadServerIfNeeded()

	ravendb.SetStateFromEnv()
	detectServerPath()

	gRavenLogsDir = ravenLogsDirFromTestName(t)

	fmt.Printf("\nStarting test %s\n", t.Name())
	var pcapPath string

	ravendb.HTTPLoggerWriter.Store(nil)
	if ravendb.LogAllRequests {
		var err error
		path := httpLogPathFromTestName(t)
		f, err := os.Create(path)
		if err != nil {
			fmt.Printf("os.Create('%s') failed with %s\n", path, err)
		} else {
			fmt.Printf("Logging HTTP traffic to %s\n", path)
		}
		ravendb.HTTPLoggerWriter.Store(io.WriteCloser(f))
	}

	ravendb.HTTPFailedRequestsLogger = nil
	if ravendb.LogFailedRequests {
		if ravendb.LogFailedRequestsDelayed {
			ravendb.HTTPFailedRequestsLogger = bytes.NewBuffer(nil)
		} else {
			ravendb.HTTPFailedRequestsLogger = os.Stdout
		}
	}

	if ravendb.PcapCapture {
		pcapPath = pcapPathFromTestName(t)
		panicIf(gRavenTestDriver != nil, "gravenTestDriver must be nil")
		gRavenTestDriver = NewRavenTestDriverWithPacketCapture(pcapPath)
	} else {
		gRavenTestDriver = NewRavenTestDriver()
	}

	return func() {
		if t.Failed() {
			maybePrintFailedRequestsLog()
		}
		deleteTestDriver()
		maybeConvertPcapToTxt(pcapPath)
		w :=ravendb.HTTPLoggerWriter.Load()
		if w != nil {
			w.(io.WriteCloser).Close()
			ravendb.HTTPLoggerWriter.Store(nil)
		}
	}
}

func recoverTest(t *testing.T, destroyDriver func()) {
	r := recover()
	destroyDriver()
	if r != nil {
		fmt.Printf("Panic: '%v'\n", r)
		debug.PrintStack()
		t.Fail()
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

func TestMain(m *testing.M) {

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
