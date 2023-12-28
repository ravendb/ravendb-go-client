package tests

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const (
	RAVENDB_TEST_PORT_START int32 = 10_000
	LOOPBACK                      = "127.0.0.1"
	LOCALHOST                     = "localhost"
)

var (
	// if 1 - no cluster
	// should be 3, 5, 7 etc.
	// can be changed via NODES_IN_CLUSTER env variable
	numClusterNodes = 1

	// a value in range 0..100
	// it's a percentage chance that we'll kill the server
	// when getting a store for a sub-test
	// 0 means killing is disabled, 5 means it's a 5% chance
	// 100 or more means it's certain
	// can be changed via KILL_SERVER_CHANCE env variable
	randomlyKillServersChance = 0

	// can be changed via SHUFFLE_CLUSTER_NODES=true env variable
	shuffleClusterNodes = false

	ravendbWindowsDownloadURL = "https://hibernatingrhinos.com/downloads/RavenDB%20for%20Windows%20x64/54000" // for local usage

	ravenWindowsZipPath = "ravendb-latest.zip"
)

type ravenProcess struct {
	cmd          *exec.Cmd
	pid          int
	stdoutReader io.ReadCloser
	stderrReader io.ReadCloser

	// auto-detected url on which to contact the server
	uri string
}

// Note: Java's RemoteTestBase is folded into RavenTestDriver
type RavenTestDriver struct {
	documentStores sync.Map // *DocumentStore => bool

	dbNameCounter int32 // atomic

	store           *ravendb.DocumentStore
	serverProcesses []*ravenProcess

	isSecure bool

	customize func(*ravendb.DatabaseRecord)

	profData    bytes.Buffer
	isProfiling bool
}

var (
	// if true, enables flaky tests
	// can be enabled by setting ENABLE_FLAKY_TESTS env variable to "true"
	enableFlakyTests = false

	// if true, enable failing tests
	// can be enabled by setting ENABLE_FAILING_TESTS env variable to "true"
	enableFailingTests   = false
	testsWereInitialized bool
	muInitializeTests    sync.Mutex

	ravendbServerExePath string

	// passed to the server as --Security.Certificate.Path
	certificatePath string

	caCertificate     *x509.Certificate
	clientCertificate *tls.Certificate

	nextPort = RAVENDB_TEST_PORT_START
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

var (
	balanceBehaviors = []ravendb.ReadBalanceBehavior{
		// ravendb.ReadBalanceBehaviorNone,
		ravendb.ReadBalanceBehaviorRoundRobin,
		ravendb.ReadBalanceBehaviorFastestNode,
	}
)

func pickRandomBalanceBehavior() ravendb.ReadBalanceBehavior {
	n := rand.Intn(len(balanceBehaviors))
	return balanceBehaviors[n]
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
		fmt.Printf("RavenDB process %d terminated \nConfiguration arguments: '%s'\n on '%s'\n", proc.pid, s, proc.uri)
	}
}

func getNextPort() int {
	n := atomic.AddInt32(&nextPort, 1)
	return int(n)
}

func startRavenServer(secure bool) (*ravenProcess, error) {
	args, err := getServerConfiguration(secure)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(ravendbServerExePath, args...)
	stdoutReader, err := cmd.StdoutPipe()
	stderrReader, err := cmd.StderrPipe()

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
		stderrReader: stderrReader,
		pid:          cmd.Process.Pid,
	}

	// parse stdout of the server to extract server listening port from line:
	// Server available on: http://127.0.0.1:50386
	wantedPrefix := "Server available on: "
	scanner := bufio.NewScanner(stdoutReader)
	startTime := time.Now()
	var outputCopy bytes.Buffer
	for scanner.Scan() {
		dur := time.Since(startTime)
		if dur > 3*time.Minute {
			break
		}
		s := scanner.Text()
		if ravenServerVerbose {
			fmt.Printf("server: %s\n", s)
		}
		outputCopy.WriteString(s + "\n")
		if !strings.HasPrefix(s, wantedPrefix) {
			continue
		}
		s = strings.TrimPrefix(s, wantedPrefix)
		proc.uri = strings.TrimSpace(s)
		break
	}
	if scanner.Err() != nil {
		fmt.Printf("startRavenServer: scanner.Err() returned '%s'\n", err)
		killServer(proc)
		return nil, scanner.Err()
	}
	if proc.uri == "" {
		s := string(outputCopy.Bytes())
		errorStr, _ := ioutil.ReadAll(stderrReader)
		fmt.Printf("startRavenServer: couldn't detect start url. Server output: %s\nError:\n%s\n", s, string(errorStr))
		killServer(proc)
		return nil, fmt.Errorf("Unable to start server")
	}
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

func getServerConfiguration(secure bool) ([]string, error) {
	var httpUrl, tcpUrl url.URL
	httpPort := getNextPort()
	tcpPort := getNextPort()
	if secure {
		httpUrl = url.URL{
			Host:   LOCALHOST + ":" + strconv.Itoa(httpPort),
			Scheme: "https",
		}
		tcpUrl = url.URL{
			Host:   LOCALHOST + ":" + strconv.Itoa(tcpPort),
			Scheme: "tcp",
		}
	} else {
		httpUrl = url.URL{
			Host:   LOOPBACK + ":" + strconv.Itoa(httpPort),
			Scheme: "http",
		}
		tcpUrl = url.URL{
			Host:   LOOPBACK + ":" + strconv.Itoa(tcpPort),
			Scheme: "tcp",
		}
	}

	args := []string{
		"--RunInMemory=true",
		"--License.Eula.Accepted=true",
		"--Setup.Mode=None",
		"--Testing.ParentProcessId=" + getProcessId(),
	}

	if secure {
		secureArgs := []string{
			"--PublicServerUrl=" + httpUrl.String(),
			"--PublicServerUrl.Tcp=" + tcpUrl.String(),
			"--ServerUrl=https://0.0.0.0:" + strconv.Itoa(httpPort),
			"--ServerUrl.Tcp=tcp://0.0.0.0:" + strconv.Itoa(tcpPort),
			"--Security.Certificate.Path=" + certificatePath,
		}
		args = append(args, secureArgs...)
	} else {
		unsecureArgs := []string{
			"--ServerUrl=" + httpUrl.String(),
			"--ServerUrl.Tcp=" + tcpUrl.String(),
		}
		args = append(args, unsecureArgs...)
	}
	return args, nil
}

func runServersMust(n int, secure bool) ([]*ravenProcess, error) {
	var wg sync.WaitGroup
	errorsChannel := make(chan error, n)
	if secure {
		n = 1
	}
	var procs []*ravenProcess
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(secureCopy bool) {
			proc, err := startRavenServer(secureCopy)
			if err != nil {
				errorsChannel <- err
			}
			args := strings.Join(proc.cmd.Args, " ")
			fmt.Printf("Started server '%s' on port '%s' pid: %d\n", args, proc.uri, proc.pid)
			procs = append(procs, proc)
			wg.Done()
		}(secure)
	}
	wg.Wait()
	close(errorsChannel)

	var result error

	for err := range errorsChannel {
		result = multierror.Append(result, err)
	}

	return procs, result
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

func (d *RavenTestDriver) customizeDbRecord(dbRecord *ravendb.DatabaseRecord) {
	if d.customize != nil {
		d.customize(dbRecord)
	}
}

func (d *RavenTestDriver) maybeKillServer() bool {
	if len(d.serverProcesses) < numClusterNodes || len(d.serverProcesses) < 2 {
		return false
	}
	// randomly kill a server
	n := rand.Intn(100)
	if n >= randomlyKillServersChance {
		return false
	}
	// don't kill the first server as it's used by main store to create
	// databases / store for other commands
	idx := 1
	proc := d.serverProcesses[idx]
	d.serverProcesses = append(d.serverProcesses[:idx], d.serverProcesses[idx+1:]...)
	fmt.Printf("Randomly killing a server with pid %d\n", proc.pid)
	killServer(proc)
	return true
}

func (d *RavenTestDriver) getDocumentStore2(dbName string, waitForIndexingTimeout time.Duration) (*ravendb.DocumentStore, error) {

	var err error

	// we're called for each sub-test
	if d.store == nil {
		d.store, err = d.createMainStore()
		if err != nil {
			return nil, err
		}
	} else {
		d.maybeKillServer()
	}

	n := int(atomic.AddInt32(&d.dbNameCounter, 1))
	name := fmt.Sprintf("%s_%d", dbName, n)
	databaseRecord := ravendb.NewDatabaseRecord()
	databaseRecord.DatabaseName = name
	d.customizeDbRecord(databaseRecord)

	// replicationFactor seems to be a minimum number of nodes with the data
	// so it must be less than 3 (we have 3 nodes and might kill one, leaving
	// only 2)
	createDatabaseOperation := ravendb.NewCreateDatabaseOperation(databaseRecord, 1)
	err = d.store.Maintenance().Server().Send(createDatabaseOperation)
	if err != nil {
		fmt.Printf("d.store.Maintenance().Server().Send(createDatabaseOperation) failed with %s\n", err)
		return nil, err
	}

	uris := d.store.GetUrls()
	var store *ravendb.DocumentStore
	if shuffleClusterNodes {
		// randomly shuffle urls so that if we kill a server, there's a higher
		// chance we'll hit it
		var shuffledURIs []string
		r := rand.New(rand.NewSource(time.Now().Unix()))
		for _, i := range r.Perm(len(uris)) {
			shuffledURIs = append(shuffledURIs, uris[i])
		}
		store = ravendb.NewDocumentStore(shuffledURIs, name)
	} else {
		store = ravendb.NewDocumentStore(uris, name)
	}
	conventions := store.GetConventions()
	conventions.ReadBalanceBehavior = pickRandomBalanceBehavior()
	store.SetConventions(conventions)

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
		fmt.Printf("getDocumentStore2: store.Initialize() failed with '%s'\n", err)
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
			fmt.Printf("getDocumentStore2:  d.waitForIndexing() failed with '%s'\n", err)
			store.Close()
			return nil, err
		}
	}

	d.documentStores.Store(store, true)
	d.maybeStartProfiling()

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

func setupCluster(store *ravendb.DocumentStore) error {
	uris := store.GetUrls()
	if len(uris) < 2 {
		return nil
	}

	re := store.GetRequestExecutor(store.GetDatabase())
	httpClient, err := re.GetHTTPClient()
	if err != nil {
		fmt.Printf("setupCluster: re.GetHTTPClient() failed with '%s'\n", err)
		return err
	}
	firstServerURL := uris[0]
	for _, uri := range uris[1:] {
		// https://ravendb.net/docs/article-page/4.1/csharp/server/clustering/cluster-api#delete-node-from-the-cluster
		cmdURI := firstServerURL + "/admin/cluster/node?url=" + url.QueryEscape(uri)
		req, err := newHttpPut(cmdURI, nil)
		if err != nil {
			fmt.Printf("setupCluster: newHttpPutt() failed with '%s'\n", err)
		}
		rsp, err := httpClient.Do(req)
		if err != nil {
			fmt.Printf("setupCluster: httpClient.Do() failed with '%s' for url '%s'\n", err, cmdURI)
		}
		defer rsp.Body.Close()
		if rsp.StatusCode >= 400 {
			fmt.Printf("setupCluster: httpClient.Do() returned status code '%s' for url '%s'\n", rsp.Status, cmdURI)
			return fmt.Errorf("setupCluster: httpClient.Do() returned status code '%s' for url '%s'\n", rsp.Status, cmdURI)
		}

		fmt.Printf("Added node to cluster with '%s', status code: %d\n", cmdURI, rsp.StatusCode)
	}
	return nil
}

func newHttpPut(uri string, data []byte) (*http.Request, error) {
	var body io.Reader
	if len(data) > 0 {
		body = bytes.NewReader(data)
	}
	req, err := http.NewRequest(http.MethodPut, uri, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("User-Agent", "ravendb-go-client/4.0.0")
	if len(data) > 0 {
		req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	}
	return req, nil
}

func (d *RavenTestDriver) createMainStore() (*ravendb.DocumentStore, error) {
	var err error
	panicIf(len(d.serverProcesses) > 0, "len(d.serverProcesses): %d", len(d.serverProcesses))

	d.serverProcesses, err = runServersMust(numClusterNodes, d.isSecure)
	if err != nil {
		return nil, err
	}

	var uris []string
	for _, proc := range d.serverProcesses {
		uris = append(uris, proc.uri)
	}

	mainStoreURLS := uris
	if len(mainStoreURLS) > 1 {
		mainStoreURLS = mainStoreURLS[1:]
	}
	store := ravendb.NewDocumentStore(uris, "test.manager")

	conventions := store.GetConventions()
	// main store is only used to create databases / other stores
	// so we don't want cluster behavior
	conventions.SetDisableTopologyUpdates(true)
	conventions.ReadBalanceBehavior = pickRandomBalanceBehavior()

	if d.isSecure {
		store.Certificate = clientCertificate
		store.TrustStore = caCertificate
	}
	err = store.Initialize()
	if err != nil {
		fmt.Printf("createMainStore: store.Initialize() failed with '%s'\n", err)
		store.Close()
		return nil, err
	}
	err = setupCluster(store)
	if err != nil {
		store.Close()
		return nil, err
	}
	return store, nil
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
	d.isSecure = false
	store, err := d.getDocumentStore2("test_db", 0)
	assert.NoError(t, err)
	assert.NotNil(t, store)
	return store
}

func (d *RavenTestDriver) getSecuredDocumentStoreMust(t *testing.T) *ravendb.DocumentStore {
	d.isSecure = true
	store, err := d.getDocumentStore2("test_db", 0)
	assert.NoError(t, err)
	assert.NotNil(t, store)
	return store
}

func (d *RavenTestDriver) Close() {
	// fmt.Print("RavenTestDriver.Close()\n")
	// debug.PrintStack()

	fn := func(key, value interface{}) bool {
		documentStore := key.(*ravendb.DocumentStore)
		documentStore.Close()
		d.documentStores.Delete(key)
		return true
	}
	d.documentStores.Range(fn)
	if d.store != nil {
		d.store.Close()
	}
	d.killServerProcesses()
}

func shutdownTests() {
	// no-op
}

func openSessionMust(t *testing.T, store *ravendb.DocumentStore) *ravendb.DocumentSession {
	session, err := store.OpenSession("")
	assert.NoError(t, err)
	assert.NotNil(t, session)
	return session
}

func openSessionMustWithOptions(t *testing.T, store *ravendb.DocumentStore, options *ravendb.SessionOptions) *ravendb.DocumentSession {
	session, err := store.OpenSessionWithOptions(options)
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

func (d *RavenTestDriver) maybeStartProfiling() {
	if !isEnvVarTrue("ENABLE_PROFILING") || d.isProfiling {
		return
	}
	if err := pprof.StartCPUProfile(&d.profData); err != nil {
		fmt.Printf("pprof.StartCPUProfile() failed with '%s'\n", err)
	} else {
		d.isProfiling = true
		fmt.Printf("started cpu profiling\n")
	}
}

func (d *RavenTestDriver) maybeStopProfiling() {
	if !d.isProfiling {
		return
	}
	pprof.StopCPUProfile()
	path := "cpu.prof"
	pd := d.profData.Bytes()
	err := ioutil.WriteFile(path, pd, 0644)
	if err != nil {
		fmt.Printf("failed to write cpu profile data to '%s'. Error: '%s'\n", path, err)
	} else {
		fmt.Printf("wrote cpu profile data to '%s'\n", path)
	}
}

// called for every Test* function
func createTestDriver(t *testing.T) *RavenTestDriver {
	fmt.Printf("\nStarting test %s\n", t.Name())
	setupLogging(t)
	driver := &RavenTestDriver{}
	return driver
}

func destroyDriver(t *testing.T, d *RavenTestDriver) {
	if t.Failed() {
		maybePrintFailedRequestsLog()
	}
	if d != nil {
		d.Close()
	}
	d.maybeStopProfiling()
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
	path := os.Getenv("RAVENDB_SERVER_PATH")

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

	{
		s := os.Getenv("NODES_IN_CLUSTER")
		n, err := strconv.Atoi(s)
		if err == nil && n > 1 {
			numClusterNodes = n
			fmt.Printf("Setting numClusterNodes=%d from NODES_IN_CLUSTER env variable\n", n)
		}
	}

	{
		s := os.Getenv("KILL_SERVER_CHANCE")
		n, err := strconv.Atoi(s)
		if err == nil {
			randomlyKillServersChance = n
			fmt.Printf("Setting randomlyKillServersChance=%d from KILL_SERVER_CHANCE env variable\n", n)
		}
	}

	if !shuffleClusterNodes && isEnvVarTrue("SHUFFLE_CLUSTER_NODES") {
		shuffleClusterNodes = true
		fmt.Printf("Setting shuffleClusterNodes to true because SHUFFLE_CLUSTER_NODES env variable is %s\n", os.Getenv("SHUFFLE_CLUSTER_NODES"))
	}

	setLoggingStateFromEnv()
	detectRavenDevLicense()

	ravendbServerExePath = detectRavendbExePath()
	if ravendbServerExePath == "" {
		if isWindows() {
			downloadServerIfNeededWindows()
			ravendbServerExePath = detectRavendbExePath()
		}
	}

	if ravendbServerExePath == "" {
		fmt.Printf("Didn't find ravendb server exe. Set RAVENDB_SERVER_PATH env variable\n")
		os.Exit(1)
	}

	// find top-level directory
	// wd should be "tests" sub-directory
	wd, _ := os.Getwd()
	rootDir := filepath.Clean(filepath.Join(wd, ".."))
	path := filepath.Join(rootDir, "certs", "server.pfx")
	if !fileExists(path) {
		fmt.Printf("rootDir '%s' doesn't seem correct because can't find file '%s'\n", rootDir, path)
		os.Exit(1)
	}

	// detect paths of files needed to run the tests
	// either get them from env variables (set by test scripts)
	// or try to auto-detect (helps running tests from within
	// Visual Studio Code or GoLand where env variables are not set)
	{
		path := os.Getenv("RAVENDB_TEST_CERTIFICATE_PATH")
		// wd should be
		if !fileExists(path) {
			path = filepath.Join(rootDir, "certs", "server.pfx")
		}
		if !fileExists(path) {
			fmt.Printf("Didn't find server.pfx file at '%s'. Set RAVENDB_TEST_CERTIFICATE_PATH env variable\n", path)
			os.Exit(1)
		}
		certificatePath = path
		fmt.Printf("Server ertificate file found at '%s'\n", certificatePath)
	}

	{
		path := os.Getenv("RAVENDB_TEST_CA_PATH")
		if !fileExists(path) {
			path = filepath.Join(rootDir, "certs", "ca.crt")
		}
		if !fileExists(path) {
			fmt.Printf("Didn't find ca.cert file at '%s'. Set RAVENDB_TEST_CA_PATH env variable\n", path)
			os.Exit(1)
		}
		caCertificate = loadTestCaCertificate(path)
		fmt.Printf("Loaded ca certificate from '%s'\n", path)
	}

	{
		path := os.Getenv("RAVENDB_TEST_CLIENT_CERTIFICATE_PATH")
		if !fileExists(path) {
			path = filepath.Join(rootDir, "certs", "go.pem")
		}
		if !fileExists(path) {
			fmt.Printf("Didn't find cert.pem file at '%s'. Set RAVENDB_TEST_CLIENT_CERTIFICATE_PATH env variable\n", path)
			os.Exit(1)
		}
		clientCertificate = loadTestClientCertificate(path)
		fmt.Printf("Loaded client certificate from '%s'\n", path)
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
