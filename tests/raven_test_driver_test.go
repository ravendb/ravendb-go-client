package tests

import (
	"bufio"
	"bytes"
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"runtime/pprof"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

var (
	// in Java those are static fields of RavenTestDriver
	globalServer               *ravendb.DocumentStore
	globalServerProcess        *Process
	globalSecuredServer        *ravendb.DocumentStore
	globalSecuredServerProcess *Process
	index                      int32
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

	disposed bool
}

func NewRavenTestDriver() *RavenTestDriver {
	return &RavenTestDriver{}
}

func parsePrivateKey(der []byte) (crypto.PrivateKey, error) {
	if key, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return key, nil
	}
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		switch key := key.(type) {
		case *rsa.PrivateKey, *ecdsa.PrivateKey:
			return key, nil
		default:
			return nil, fmt.Errorf("Found unknown private key type in PKCS#8 wrapping")
		}
	}
	if key, err := x509.ParseECPrivateKey(der); err == nil {
		return key, nil
	}
	return nil, fmt.Errorf("Failed to parse private key")
}

func loadCertficateAndKeyFromFile(path string) (*tls.Certificate, error) {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cert tls.Certificate
	for {
		block, rest := pem.Decode(raw)
		if block == nil {
			break
		}
		if block.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, block.Bytes)
		} else {
			cert.PrivateKey, err = parsePrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("Failure reading private key from \"%s\": %s", path, err)
			}
		}
		raw = rest
	}

	if len(cert.Certificate) == 0 {
		return nil, fmt.Errorf("No certificate found in \"%s\"", path)
	} else if cert.PrivateKey == nil {
		return nil, fmt.Errorf("No private key found in \"%s\"", path)
	}

	return &cert, nil
}

func getTestClientCertificate() *ravendb.KeyStore {
	res := &ravendb.KeyStore{}
	path := os.Getenv("RAVENDB_JAVA_TEST_CLIENT_CERTIFICATE_PATH")
	cert, err := loadCertficateAndKeyFromFile(path)
	must(err)
	res.Certificates = append(res.Certificates, *cert)
	return res
}

func (d *RavenTestDriver) getDocumentStore() (*ravendb.DocumentStore, error) {
	return d.getDocumentStore2("test_db", false, 0)
}

func (d *RavenTestDriver) getSecuredDocumentStore() (*ravendb.DocumentStore, error) {
	return d.getDocumentStore2("test_db", true, 0)
}

func (d *RavenTestDriver) getDocumentStore2(dbName string, secured bool, waitForIndexingTimeout time.Duration) (*ravendb.DocumentStore, error) {

	n := int(atomic.AddInt32(&index, 1))
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
		store.SetCertificate(getTestClientCertificate())
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

	if ravendb.RavenServerVerbose {
		go func() {
			_, err = io.Copy(os.Stdout, proc.stdoutReader)
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
		globalSecuredServer = store
		clientCert := getTestClientCertificate()
		store.SetCertificate(clientCert)
	} else {
		globalServer = store
	}
	err = store.Initialize()
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
			if index.State == ravendb.IndexState_DISABLED {
				continue
			}
			if index.IsStale || strings.HasPrefix(index.Name, ravendb.Constants_Documents_Indexing_SIDE_BY_SIDE_INDEX_NAME_PREFIX) {
				isDone = false
			}
			if index.State == ravendb.IndexState_ERROR {
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
}

var dbTestsDisabledAlreadyPrinted = false

func isEnvVarTrue(name string) bool {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(name)))
	switch v {
	case "yes", "true":
		return true
	}
	return false
}

func dbTestsDisabled() bool {
	if isEnvVarTrue("RAVEN_GO_NO_DB_TESTS") {
		if !dbTestsDisabledAlreadyPrinted {
			dbTestsDisabledAlreadyPrinted = true
			fmt.Printf("DB tests are disabled\n")
		}
		return true
	}
	return false
}

func getDocumentStoreMust(t *testing.T, driver *RavenTestDriver) *ravendb.DocumentStore {
	store, err := driver.getDocumentStore()
	assert.NoError(t, err)
	assert.NotNil(t, store)
	return store
}

func getSecuredDocumentStoreMust(t *testing.T, driver *RavenTestDriver) *ravendb.DocumentStore {
	store, err := driver.getSecuredDocumentStore()
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
	os.MkdirAll(dir, 0755)
	return dir
}

func httpLogPathFromTestName(t *testing.T) string {
	name := "trace_" + testNameToFileName(t.Name()) + "_go.txt"
	return filepath.Join(getLogDir(), name)
}

func ravenLogsDirFromTestName(t *testing.T) string {
	name := testNameToFileName(t.Name())
	path := filepath.Join(getLogDir(), "server", "go", name)
	// recreate dir for clean logs
	os.RemoveAll(path)
	os.MkdirAll(path, 0755)
	return path
}

func deleteTestDriver(driver *RavenTestDriver) {
	if driver == nil {
		return
	}
	driver.Close()
	killGlobalServerProcesses()
}

var (
	defaultUserAgent          = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/70.0.3538.77 Safari/537.36"
	ravendbWindowsDownloadURL = "https://daily-builds.s3.amazonaws.com/RavenDB-4.1.2-windows-x64.zip"
	ravenWindowsZipPath       = "Ravendb-4.1.2.zip"
)

func getRavendbExePath() string {
	cwd, err := os.Getwd()
	must(err)

	path := filepath.Join(cwd, "..", "RavenDB", "Server", "Raven.Server")
	if isWindows() {
		path += ".exe"
	}
	if ravendb.FileExists(path) {
		return path
	}

	path = filepath.Join(cwd, "RavenDB", "Server", "Raven.Server")
	if isWindows() {
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
		err = HttpDl(ravendbWindowsDownloadURL, ravenWindowsZipPath)
		must(err)
		fmt.Printf(" took %s\n", time.Since(timeStart))
	}
	destDir := "RavenDB"
	fmt.Printf("Unzipping %s to %s...", ravenWindowsZipPath, destDir)
	timeStart := time.Now()
	err = Unzip(ravenWindowsZipPath, destDir)
	must(err)
	fmt.Printf(" took %s\n", time.Since(timeStart))
}

var muServerDownload sync.Mutex

func downloadServerIfNeeded() {
	muServerDownload.Lock()
	defer muServerDownload.Unlock()
	if isWindows() {
		downloadServerIfNeededWindows()
		return
	}
}

// this helps running tests from within Visual Studio Code,
// where env variables are not set
func detectServerPath() {
	// explicitly setting RAVEN_GO_NO_DB_TESTS=true disables database tests
	// so no need for the server
	if isEnvVarTrue("RAVEN_GO_NO_DB_TESTS") {
		return
	}

	// auto-detect env variables if not explicitly set
	serverPath := os.Getenv("RAVENDB_JAVA_TEST_SERVER_PATH")
	if serverPath == "" {
		path := getRavendbExePath()
		_, err := os.Stat(path)
		must(err)
		os.Setenv("RAVENDB_JAVA_TEST_SERVER_PATH", path)
		fmt.Printf("Setting RAVENDB_JAVA_TEST_SERVER_PATH to '%s'\n", path)
	}

	if os.Getenv("RAVENDB_JAVA_TEST_CERTIFICATE_PATH") == "" {
		path := filepath.Join("..", "certs", "server.pfx")
		_, err := os.Stat(path)
		must(err)
		os.Setenv("RAVENDB_JAVA_TEST_CERTIFICATE_PATH", path)
		fmt.Printf("Setting RAVENDB_JAVA_TEST_CERTIFICATE_PATH to '%s'\n", path)
	}

	if os.Getenv("RAVENDB_JAVA_TEST_CLIENT_CERTIFICATE_PATH") == "" {
		path := filepath.Join("..", "certs", "cert.pem")
		_, err := os.Stat(path)
		must(err)
		os.Setenv("RAVENDB_JAVA_TEST_CLIENT_CERTIFICATE_PATH", path)
		fmt.Printf("Setting RAVENDB_JAVA_TEST_CLIENT_CERTIFICATE_PATH to '%s'\n", path)
	}

	if os.Getenv("RAVENDB_JAVA_TEST_HTTPS_SERVER_URL") == "" {
		uri := "https://a.javatest11.development.run:8085"
		os.Setenv("RAVENDB_JAVA_TEST_HTTPS_SERVER_URL", uri)
		fmt.Printf("Setting RAVENDB_JAVA_TEST_HTTPS_SERVER_URL to '%s'\n", uri)
	}
}

func maybePrintFailedRequestsLog() {
	if ravendb.LogFailedRequests && ravendb.LogFailedRequestsDelayed {
		buf := ravendb.HTTPFailedRequestsLogger.(*bytes.Buffer)
		os.Stdout.Write(buf.Bytes())
		buf.Reset()
	}
}

// for temporarily disabling logging of failed requests (if a given
// test is known to issue failing requests)
// usage: defer disableLogFailedRequests()()
// or:
// restorer := disableLogFailedRequests()
// ...
// restorer()
func disableLogFailedRequests() func() {
	old := ravendb.LogFailedRequests
	ravendb.LogFailedRequests = false
	return func() {
		ravendb.LogFailedRequests = old
	}
}

// In Java, RavenTestDriver is created/destroyed for each test
// In Go we have to do it manually
// returns a shutdown function that must be called to cleanly shutdown test
func createTestDriver(t *testing.T) *RavenTestDriver {
	downloadServerIfNeeded()

	ravendb.SetStateFromEnv()
	detectServerPath()

	gRavenLogsDir = ravenLogsDirFromTestName(t)

	fmt.Printf("\nStarting test %s\n", t.Name())

	ravendb.LogsLock()
	defer ravendb.LogsUnlock()

	ravendb.HTTPLoggerWriter = nil
	if ravendb.LogAllRequests {
		var err error
		path := httpLogPathFromTestName(t)
		f, err := os.Create(path)
		if err != nil {
			fmt.Printf("os.Create('%s') failed with %s\n", path, err)
		} else {
			fmt.Printf("Logging HTTP traffic to %s\n", path)
			ravendb.HTTPLoggerWriter = f
		}
	}

	ravendb.HTTPFailedRequestsLogger = nil
	if ravendb.LogFailedRequests {
		if ravendb.LogFailedRequestsDelayed {
			ravendb.HTTPFailedRequestsLogger = bytes.NewBuffer(nil)
		} else {
			ravendb.HTTPFailedRequestsLogger = os.Stdout
		}
	}

	return NewRavenTestDriver()
}

func destroyDriver(t *testing.T, driver *RavenTestDriver) {
	if t.Failed() {
		maybePrintFailedRequestsLog()
	}
	deleteTestDriver(driver)
	ravendb.LogsLock()
	defer ravendb.LogsUnlock()
	w := ravendb.HTTPLoggerWriter
	if w != nil {
		w.Close()
		ravendb.HTTPLoggerWriter = nil
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

func isWindows() bool {
	return runtime.GOOS == "windows"
}

func timeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		conn.SetDeadline(time.Now().Add(rwTimeout))
		return conn, nil
	}
}

// can be used for http.Get() requests with better timeouts. New one must be created
// for each Get() request
func newTimeoutClient(connectTimeout time.Duration, readWriteTimeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			Dial:  timeoutDialer(connectTimeout, readWriteTimeout),
			Proxy: http.ProxyFromEnvironment,
		},
	}
}

func downloadURL(url string) ([]byte, error) {
	// default timeout for http.Get() is really long, so dial it down
	// for both connection and read/write timeouts
	timeoutClient := newTimeoutClient(time.Second*120, time.Second*120)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", defaultUserAgent)
	resp, err := timeoutClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("'%s': status code not 200 (%d)", url, resp.StatusCode)
	}
	return ioutil.ReadAll(resp.Body)
}

func HttpDl(url string, destPath string) error {
	d, err := downloadURL(url)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(destPath, d, 0755)
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
