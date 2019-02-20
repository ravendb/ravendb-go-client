package tests

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

type ravenProcess struct {
	cmd          *exec.Cmd
	stdoutReader io.ReadCloser

	// auto-detected url on which to contact the server
	uri string
}

// Note: Java's RemoteTestBase is folded into RavenTestDriver
type RavenTestDriver struct {
	documentStores sync.Map // *DocumentStore => bool

	dbNameCounter int32 // atomic

	store         *ravendb.DocumentStore
	serverProcesses []*ravenProcess

	isSecure bool

	disposed bool

	customize func(*ravendb.DatabaseRecord)
}

var (
	// if true, enables flaky tests
	// can be enabled by setting ENABLE_FLAKY_TESTS env variable to "true"
	enableFlakyTests = false

	// if true, enable failing tests
	// can be enabled by setting ENABLE_FAILING_TESTS env variable to "true"
	enableFailingTests = false

	testsWereInitialized bool
	muInitializeTests sync.Mutex

	ravendbServerExePath string

	// passed to the server as --Security.Certificate.Path
	certificatePath string

	caCertificate *x509.Certificate
	clientCertificate *tls.Certificate

	httpsServerURL string

	tcpServerPort int32 = 38880// atomic
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

func killServer(proc *ravenProcess) {
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
}

func getNextTcpPort() int {
	n := atomic.AddInt32(&tcpServerPort, 1)
	return int(n)
}

func startRavenServer(secure bool) (*ravenProcess, error) {
	serverURL := "http://127.0.0.1:0"
	// we run potentially multiple server so need to make the port unique
	tcpServerURL := fmt.Sprintf("tcp://127.0.0.1:%d", getNextTcpPort())

	if secure {
		serverURL = httpsServerURL
		parsed, err := url.Parse(httpsServerURL)
		must(err)
		host := parsed.Host
		parts := strings.Split(host, ":")
		panicIf(len(parts) > 2, "invalid https URL '%s'", httpsServerURL)
		// host can be name:port, extract "name" part
		host = parts[0]
		tcpServerURL = "tcp://" + host + ":38882"
	}

	args := []string{
		"--ServerUrl=" + serverURL,
		"--ServerUrl.Tcp=" + tcpServerURL,
		"--RunInMemory=true",
		"--License.Eula.Accepted=true",
		"--Setup.Mode=None",
		"--Testing.ParentProcessId=" + getProcessId(),
		// "--non-interactive",
	}

	if secure {
		secureArgs := []string{
			"--Security.Certificate.Path=" + certificatePath,
			"--Security.Certificate.Password=pwd1234",
		}
		args = append(args, secureArgs...)
	}

	cmd := exec.Command(ravendbServerExePath, args...)
	stdoutReader, err := cmd.StdoutPipe()

	if false && ravenServerVerbose {
		cmd.Stderr = os.Stderr
		// cmd.StdoutPipe() sets cmd.Stdout to a pipe writer
		// we multi-plex it into os.Stdout
		// TODO: this doesn't seem to work. It makes reading from stdoutReader
		// immediately fail. Maybe it's becuse writer returned by
		// os.Pipe() (cmd.Stdout) blocks and MultiWriter() doesn't
		cmd.Stdout = io.MultiWriter(cmd.Stdout, os.Stdout)
	}
	if err != nil {
		return nil, err
	}
	err = cmd.Start()
	if err != nil {
		fmt.Printf("exec.Command(%s, %v) failed with %s\n", ravendbServerExePath, args, err)
		return nil, err
	}

	proc := &ravenProcess{
		cmd:          cmd,
		stdoutReader: stdoutReader,
	}

	// parse stdout of the server to extract server listening port from line:
	// Server available on: http://127.0.0.1:50386
	wantedPrefix := "Server available on: "
	scanner := bufio.NewScanner(stdoutReader)
	startTime := time.Now()
	for scanner.Scan() {
		dur := time.Since(startTime)
		if dur > time.Minute {
			break
		}
		s := scanner.Text()
		if ravenServerVerbose {
			fmt.Printf("%s\n", s)
		}
		if !strings.HasPrefix(s, wantedPrefix) {
			continue
		}
		s = strings.TrimPrefix(s, wantedPrefix)
		proc.uri = strings.TrimSpace(s)
		break
	}
	if scanner.Err() != nil {
		killServer(proc)
		return nil, scanner.Err()
	}
	if proc.uri == "" {
		killServer(proc)
		return nil, fmt.Errorf("Unable to start server")
	}
	fmt.Printf("Server started on: '%s'\n", proc.uri)

	if ravenServerVerbose {
		go func() {
			_, err = io.Copy(os.Stdout, stdoutReader)
			if !(err == nil || err == io.EOF) {
				fmt.Printf("io.Copy() failed with %s\n", err)
			}
		}()
	}

	time.Sleep(time.Millisecond * 100) // TODO: probably not necessary

	return proc, nil
}

func setupRevisions(store *ravendb.DocumentStore, purgeOnDelete bool, minimumRevisionsToKeep int64) (*ravendb.ConfigureRevisionsOperationResult, error) {

	revisionsConfiguration := &ravendb.RevisionsConfiguration{}
	defaultCollection := &ravendb.RevisionsCollectionConfiguration{}
	defaultCollection.PurgeOnDelete = purgeOnDelete
	defaultCollection.MinimumRevisionsToKeep = minimumRevisionsToKeep

	revisionsConfiguration.DefaultConfig = defaultCollection
	operation := ravendb.NewConfigureRevisionsOperation(revisionsConfiguration)

	err := store.Maintenance().Send(operation)
	if err != nil {
		return nil, err
	}

	return operation.Command.Result, nil
}

func (d *RavenTestDriver) getDocumentStore() (*ravendb.DocumentStore, error) {
	d.isSecure = false
	return d.getDocumentStore2("test_db",  0)
}

func (d *RavenTestDriver) getSecuredDocumentStore() (*ravendb.DocumentStore, error) {
	d.isSecure = false
	return d.getDocumentStore2("test_db",  0)
}

func (d *RavenTestDriver) customizeDbRecord(dbRecord *ravendb.DatabaseRecord) {
	if d.customize != nil {
		d.customize(dbRecord)
	}
}
func (d *RavenTestDriver) getDocumentStore2(dbName string, waitForIndexingTimeout time.Duration) (*ravendb.DocumentStore, error) {

	n := int(atomic.AddInt32(&d.dbNameCounter, 1))
	name := fmt.Sprintf("%s_%d", dbName, n)
	documentStore := d.store
	if documentStore == nil {
		err := d.runServer()
		if err != nil {
			fmt.Printf("runServer failed with %s\n", err)
			return nil, err
		}
	}

	documentStore = d.store
	databaseRecord := ravendb.NewDatabaseRecord()
	databaseRecord.DatabaseName = name

	d.customizeDbRecord(databaseRecord)

	createDatabaseOperation := ravendb.NewCreateDatabaseOperation(databaseRecord)
	err := documentStore.Maintenance().Server().Send(createDatabaseOperation)
	if err != nil {
		return nil, err
	}

	urls := documentStore.GetUrls()
	store := ravendb.NewDocumentStore(urls, name)

	if d.isSecure {
		store.Certificate = clientCertificate
		store.TrustStore = caCertificate
	}

	// TODO: is over-written by CustomSerializationTest
	// customizeStore(Store);
	d.hookLeakedConnectionCheck(store)

	d.setupDatabase(store)
	err = store.Initialize()
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
		err = store.Maintenance().Server().Send(operation)
	}

	store.AddAfterCloseListener(fn)

	if waitForIndexingTimeout > 0 {
		err = d.waitForIndexing(store, name, waitForIndexingTimeout)
		if err != nil {
			store.Close()
			return nil, err
		}
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

func (d *RavenTestDriver) runServer() error {
	nServers := 3
	if d.isSecure {
		nServers = 1
	}
	for i := 0; i < nServers; i++ {
		proc, err := startRavenServer(d.isSecure)
		if err != nil {
			fmt.Printf("startRavenServer failed with %s\n", err)
			return err
		} else {
			args := strings.Join(proc.cmd.Args, " ")
			fmt.Printf("Started raven server '%s'\n", args)
		}
		d.serverProcesses = append(d.serverProcesses, proc)
	}

	var uris []string
	for _, proc := range d.serverProcesses {
		uris = append(uris, proc.uri)
	}

	store := ravendb.NewDocumentStore(nil, "")
	store.SetDatabase("test.manager")
	store.SetUrls(uris)
	store.GetConventions().SetDisableTopologyUpdates(true)
	d.store = store

	if d.isSecure {
		store.Certificate = clientCertificate
		store.TrustStore = caCertificate
	}
	err := store.Initialize()
	return err
}

func (d *RavenTestDriver) waitForIndexing(store *ravendb.DocumentStore, database string, timeout time.Duration) error {
	return waitForIndexing(store, database, timeout)
}

func waitForIndexing(store *ravendb.DocumentStore, database string, timeout time.Duration) error {
	admin := store.Maintenance().ForDatabase(database)
	if timeout == 0 {
		timeout = time.Minute
	}

	sp := time.Now()
	for time.Since(sp) < timeout {
		op := ravendb.NewGetStatisticsOperation("")
		err := admin.Send(op)
		if err != nil {
			return err
		}
		databaseStatistics := op.Command.Result
		isDone := true
		hasError := false
		for _, index := range databaseStatistics.Indexes {
			if index.State == ravendb.IndexStateDisabled {
				continue
			}
			if index.IsStale || strings.HasPrefix(index.Name, ravendb.IndexingSideBySideIndexNamePrefix) {
				isDone = false
			}
			if index.State == ravendb.IndexStateError {
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
	return ravendb.NewTimeoutError("The indexes stayed stale for more than %s.%s", timeout, allIndexErrorsText)
}

func (d *RavenTestDriver) killServerProcesses() {
	for _, proc := range d.serverProcesses {
		killServer(proc)
	}
	d.serverProcesses = nil
	d.store = nil
}

func (d *RavenTestDriver) getDocumentStoreMust(t *testing.T) *ravendb.DocumentStore {
	store, err := d.getDocumentStore()
	assert.NoError(t, err)
	assert.NotNil(t, store)
	return store
}

func (d *RavenTestDriver) getSecuredDocumentStoreMust(t *testing.T) *ravendb.DocumentStore {
	store, err := d.getSecuredDocumentStore()
	assert.NoError(t, err)
	assert.NotNil(t, store)
	return store
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
}

func openSessionMust(t *testing.T, store *ravendb.DocumentStore) *ravendb.DocumentSession {
	session, err := store.OpenSession("")
	assert.NoError(t, err)
	assert.NotNil(t, session)
	return session
}

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

func getLogDir() string {
	// if this is not full path, raven will put it in it's own Logs directory
	// next to server executable
	cwd, _ := os.Getwd()
	dir, file := filepath.Split(cwd)
	if file != "tests" {
		dir = cwd
	}
	dir = filepath.Join(dir, "logs")
	_ = os.MkdirAll(dir, 0755)
	return dir
}

func httpLogPathFromTestName(t *testing.T) string {
	name := "trace_" + testNameToFileName(t.Name()) + "_go.txt"
	return filepath.Join(getLogDir(), name)
}

var (
	defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36"
	// ravendbWindowsDownloadURL = "https://daily-builds.s3.amazonaws.com/RavenDB-4.1.3-windows-x64.zip"
	ravendbWindowsDownloadURL = "https://hibernatingrhinos.com/downloads/RavenDB%20for%20Windows%20x64/latest?buildType=nightly&version=4.1"

	ravenWindowsZipPath = "ravendb-latest.zip"
)

// called for every Test* function
func createTestDriver(t *testing.T) *RavenTestDriver {
	fmt.Printf("\nStarting test %s\n", t.Name())
	setupLogging(t)
	driver := &RavenTestDriver{}
	return driver
}

func destroyDriver(t *testing.T, driver *RavenTestDriver) {
	if t.Failed() {
		maybePrintFailedRequestsLog()
	}
	if driver != nil {
		driver.Close()
		driver.killServerProcesses()
	}

	finishLogging()
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

func downloadServerIfNeededWindows() {
	// hacky: if we're in tests directory, cd .. for duration of this function
	panicIf(ravendbServerExePath != "", "ravendb exe already found in %s", ravendbServerExePath)

	cwd, err := os.Getwd()
	must(err)
	if strings.HasSuffix(cwd, "tests") {
		path := filepath.Clean(filepath.Join(cwd, ".."))
		err = os.Chdir(path)
		must(err)
		defer func() {
			err := os.Chdir(cwd)
			must(err)
		}()
	}

	exists := fileExists(ravenWindowsZipPath)
	if !exists {
		fmt.Printf("Downloading %s...", ravendbWindowsDownloadURL)
		startTime := time.Now()
		err = httpDl(ravendbWindowsDownloadURL, ravenWindowsZipPath)
		must(err)
		fmt.Printf(" took %s\n", time.Since(startTime))
	}
	destDir := "RavenDB"
	fmt.Printf("Unzipping %s to %s...", ravenWindowsZipPath, destDir)
	startTime := time.Now()
	err = unzip(ravenWindowsZipPath, destDir)
	must(err)
	fmt.Printf(" took %s\n", time.Since(startTime))
}

func detectRavendbExePath() string {
	// auto-detect env variables if not explicitly set
	path := os.Getenv("RAVENDB_JAVA_TEST_SERVER_PATH")

	defer func() {
		if path != "" {
			fmt.Printf("Server exe: %s\n", path)
		}
	}()

	if fileExists(path) {
		return path
	}

	cwd, err := os.Getwd()
	must(err)

	path = filepath.Join(cwd, "..", "RavenDB", "Server", "Raven.Server")
	if isWindows() {
		path += ".exe"
	}
	path = filepath.Clean(path)
	if fileExists(path) {
		return path
	}

	path = filepath.Join(cwd, "RavenDB", "Server", "Raven.Server")
	if isWindows() {
		path += ".exe"
	}
	path = filepath.Clean(path)
	if fileExists(path) {
		return path
	}
	return ""
}

func loadTestClientCertificate(path string) *tls.Certificate {
	cert, err := loadCertficateAndKeyFromFile(path)
	must(err)
	return cert
}

func loadTestCaCertificate(path string) *x509.Certificate {
	certPEM, err := ioutil.ReadFile(path)
	must(err)
	block, _ := pem.Decode([]byte(certPEM))
	panicIf(block == nil, "failed to decode cert PEM from %s", path)
	cert, err := x509.ParseCertificate(block.Bytes)
	must(err)
	return cert
}
// for CI we set RAVEN_License env variable to dev license, so that
// we can run replication tests. On local machines I have dev license
// as a file raven_license.json
func detectRavenDevLicense() {
	if len(os.Getenv("RAVEN_License")) > 0 {
		fmt.Print("RAVEN_License env variable is set\n")
		return
	}

	path := os.Getenv("RAVEN_License_Path")
	cwd, err := os.Getwd()
	must(err)
	if !fileExists(path) {
		path = filepath.Clean(filepath.Join(cwd, "..", "raven_license.json"))
		if !fileExists(path) {
			path = filepath.Clean(filepath.Join(cwd, "..", "..", "raven_license.json"))
			if !fileExists(path) {
				fmt.Printf("Replication tests are disabled because RAVEN_License_Path not set and file %s doesn't exist.\n", path)
				return
			}
		}
		_ = os.Setenv("RAVEN_License_Path", path)
		fmt.Printf("Setting RAVEN_License_Path to '%s'\n", path)
	}
}

// note: in Java for tests marked as @DisabledOn41Server
func isRunningOn41Server() bool {
	v := os.Getenv("RAVENDB_SERVER_VERSION")
	return strings.HasPrefix(v, "4.1")
}

func initializeTests() {
	muInitializeTests.Lock()
	defer muInitializeTests.Unlock()
	if testsWereInitialized {
		return
	}

	if !enableFlakyTests && isEnvVarTrue("ENABLE_FLAKY_TESTS") {
		enableFlakyTests = true
		fmt.Printf("Setting enableFlakyTests to true\n")
	}

	if !enableFailingTests && isEnvVarTrue("ENABLE_FAILING_TESTS") {
		enableFailingTests = true
		fmt.Printf("Setting enableFailingTests to true\n")
	}

	setLoggingStateFromEnv()
	detectRavenDevLicense()

	ravendbServerExePath = detectRavendbExePath()
	if ravendbServerExePath == "" {
		if isWindows() {
			downloadServerIfNeededWindows()
		}
	}

	if ravendbServerExePath == "" {
		fmt.Printf("Didn't find ravendb server exe. Set RAVENDB_JAVA_TEST_SERVER_PATH env variable\n")
		os.Exit(1)
	}

	// detect paths of files needed to run the tests
	// either get them from env variables (set by test scripts)
	// or try to auto-detect (helps running tests from within
	// Visual Studio Code or GoLand where env variables are not set)
	{
		path := os.Getenv("RAVENDB_JAVA_TEST_CERTIFICATE_PATH")
		if !fileExists(path) {
			path = filepath.Join("..", "certs", "server.pfx")
		}
		if !fileExists(path) {
			fmt.Printf("Didn't find server.pfx file at '%s'. Set RAVENDB_JAVA_TEST_CERTIFICATE_PATH env variable\n", path)
			os.Exit(1)
		}
		certificatePath = path
		fmt.Printf("Server ertificate file found at '%s'\n", certificatePath)
	}

	{
		path := os.Getenv("RAVENDB_JAVA_TEST_CA_PATH")
		if !fileExists(path) {
			path = filepath.Join("..", "certs", "ca.crt")
		}
		if !fileExists(path) {
			fmt.Printf("Didn't find ca.cert file at '%s'. Set RAVENDB_JAVA_TEST_CA_PATH env variable\n", path)
			os.Exit(1)
		}
		caCertificate = loadTestCaCertificate(path)
		fmt.Printf("Loaded ca certificate from '%s'\n", path)
	}

	{
		path := os.Getenv("RAVENDB_JAVA_TEST_CLIENT_CERTIFICATE_PATH")
		if !fileExists(path) {
			path = filepath.Join("..", "certs", "cert.pem")
		}
		if !fileExists(path) {
			fmt.Printf("Didn't find cert.pem file at '%s'. Set RAVENDB_JAVA_TEST_CLIENT_CERTIFICATE_PATH env variable\n", path)
			os.Exit(1)
		}
		clientCertificate = loadTestClientCertificate(path)
		fmt.Printf("Loaded client certificate from '%s'\n", path)
	}

	{
		uri := os.Getenv("RAVENDB_JAVA_TEST_HTTPS_SERVER_URL")
		if uri == "" {
			uri = "https://a.javatest11.development.run:8085"
		}
		httpsServerURL = uri
		fmt.Printf("HTTPS url: '%s'\n", httpsServerURL)
	}

	testsWereInitialized = true
}

func TestMain(m *testing.M) {

	//ravenServerVerbose = true

	var code int

	// make sure it's called even if panic happens
	defer func() {
		shutdownTests()

		//logGoroutines()
		os.Exit(code)
	}()

	initializeTests()

	code = m.Run()
}
