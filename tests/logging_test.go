package tests

import (
	"bytes"
	"fmt"
	"github.com/ravendb/ravendb-go-client"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"path/filepath"
	"runtime/pprof"
	"sync"
	"sync/atomic"
	"testing"
)

var (
	// if true, printing of failed reqeusts is delayed until PrintFailedRequests
	// is called.
	// can be enabled by setting LOG_FAILED_HTTP_REQUESTS_DELAYED env variable to "true"
	logFailedRequestsDelayed = false

	// if true, logs all http requests/responses to a file for further inspection
	// this is for use in tests so the file has a fixed location:
	// logs/trace_${test_name}_go.txt
	logAllRequests = false

	// if true, we log RavenDB's output to stdout
	// can be enabled by setting LOG_RAVEN_SERVER env variable to "true"
	ravenServerVerbose = false

	// if true, logs summary of all HTTP requests i.e. "GET /foo" to stdout
	logRequestSummary = false

	// if true, logs request and response of failed http requests (i.e. those returning
	// status code >= 400) to stdout
	logFailedRequests = false

	// httpLoggerWriter is where we log all http requests and responses
	httpLoggerWriter io.WriteCloser
	// httpFailedRequestsLogger is where we log failed http requests.
	// it's either os.Stdout for immediate logging or bytes.Buffer for delayed logging
	httpFailedRequestsLogger io.Writer
	// httpRequestCount numbers http requests which helps to match http
	// traffic from java client with go client
	httpRequestCount int32

	errLogDisabled int32 // atomic, if > 0, don't log error requests

	muLog sync.Mutex
)

func logsLock() {
	muLog.Lock()
}

func logsUnlock() {
	muLog.Unlock()
}

func setLoggingStateFromEnv() {
	if !ravenServerVerbose && isEnvVarTrue("LOG_RAVEN_SERVER") {
		ravenServerVerbose = true
		fmt.Printf("Setting ravenServerVerbose to true\n")
	}

	if !logFailedRequestsDelayed && isEnvVarTrue("LOG_FAILED_HTTP_REQUESTS_DELAYED") {
		logFailedRequestsDelayed = true
		fmt.Printf("Setting logFailedRequestsDelayed to true\n")
	}

	if !ravendb.LogVerbose && isEnvVarTrue("VERBOSE_LOG") {
		ravendb.LogVerbose = true
		fmt.Printf("Setting LogVerbose to true\n")
	}

	if !logRequestSummary && isEnvVarTrue("LOG_HTTP_REQUEST_SUMMARY") {
		logRequestSummary = true
		fmt.Printf("Setting LogRequestSummary to true\n")
	}

	if !logFailedRequests && isEnvVarTrue("LOG_FAILED_HTTP_REQUESTS") {
		logFailedRequests = true
		fmt.Printf("Setting LogFailedRequests to true\n")
	}

	if !logAllRequests && isEnvVarTrue("LOG_ALL_REQUESTS") {
		logAllRequests = true
		fmt.Printf("Setting logAllRequests to true\n")
	}
}

type loggingTransport struct {
	originalTransport http.RoundTripper
}

func (t *loggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	atomic.AddInt32(&httpRequestCount, 1)

	maybeLogRequestSummary(req)
	maybeCaptureRequestBody(req)
	rsp, err := t.originalTransport.RoundTrip(req)
	maybeLogFailedResponse(req, rsp, err)
	maybeLogHTTPRequest(req, rsp, err)
	return rsp, err
}

func httpClientProcessor(c *http.Client) {
	t := c.Transport
	c.Transport = &loggingTransport{
		originalTransport: t,
	}
}

func setupLogging(t *testing.T) {
	logsLock()
	defer logsUnlock()

	ravendb.HTTPClientPostProcessor = httpClientProcessor

	httpLoggerWriter = nil
	if logAllRequests {
		var err error
		path := httpLogPathFromTestName(t)
		f, err := os.Create(path)
		if err != nil {
			fmt.Printf("os.Create('%s') failed with %s\n", path, err)
		} else {
			fmt.Printf("Logging HTTP traffic to %s\n", path)
			httpLoggerWriter = f
		}
	}

	httpFailedRequestsLogger = nil
	if logFailedRequests {
		if logFailedRequestsDelayed {
			httpFailedRequestsLogger = bytes.NewBuffer(nil)
		} else {
			httpFailedRequestsLogger = os.Stdout
		}
	}
}

func finishLogging() {
	logsLock()
	defer logsUnlock()
	w := httpLoggerWriter
	if w != nil {
		w.Close()
		httpLoggerWriter = nil
	}
}

func isErrLoggingDisabled() bool {
	n := atomic.LoadInt32(&errLogDisabled)
	return n > 0
}

// for temporarily disabling logging of failed requests (if a given
// test is known to issue failing requests)
// usage: defer disableLogFailedRequests()()
// or:
// restorer := disableLogFailedRequests()
// ...
// restorer()
// this is not perfect in parallel tests because (it might over-disable)
// but we're not doing parallel tests
func disableLogFailedRequests() func() {
	atomic.AddInt32(&errLogDisabled, 1)
	return func() {
		atomic.AddInt32(&errLogDisabled, -1)
	}
}

// returns copy of resp.Body but also makes it available for subsequent reads
func getCopyOfResponseBody(resp *http.Response) ([]byte, error) {
	if resp == nil {
		return nil, nil
	}
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = ioutil.NopCloser(bytes.NewBuffer(d))
	return d, nil
}

func logRequestAndResponseToWriter(w io.Writer, req *http.Request, rsp *http.Response, reqErr error) {
	n := atomic.LoadInt32(&httpRequestCount)

	fmt.Fprintf(w, "=========== %d:\n", n)
	if reqErr != nil {
		fmt.Fprintf(w, "%s\n", reqErr)
	}

	d, err := httputil.DumpRequest(req, false)
	if err == nil {
		w.Write(d)
	}

	if req.Body != nil {
		if cr, ok := req.Body.(*CapturingReadCloser); ok {
			body := cr.capturedData.Bytes()
			if len(body) > 0 {
				fmt.Fprintf(w, "Request body %d bytes:\n%s\n", len(body), maybePrettyPrintJSON(body))
			}
		} else {
			fmt.Fprint(w, "Can't get request body\n")
		}
	}

	if reqErr != nil {
		return
	}

	if rsp == nil {
		fmt.Fprint(w, "No response\n")
		return
	}
	fmt.Fprint(w, "--------\n")
	d, err = httputil.DumpResponse(rsp, false)
	if err == nil {
		w.Write(d)
	}
	if d, err := getCopyOfResponseBody(rsp); err != nil {
		fmt.Fprintf(w, "Failed to read response body. Error: '%s'\n", err)
	} else {
		if len(d) > 0 {
			fmt.Fprintf(w, "Response body %d bytes:\n%s\n", len(d), maybePrettyPrintJSON(d))
		}
	}
}

func maybePrintFailedRequestsLog() {
	logsLock()
	defer logsUnlock()
	if logFailedRequests && logFailedRequestsDelayed {
		buf := httpFailedRequestsLogger.(*bytes.Buffer)
		os.Stdout.Write(buf.Bytes())
		buf.Reset()
	}
}

func maybeLogHTTPRequest(req *http.Request, rsp *http.Response, err error) {
	logsLock()
	defer logsUnlock()

	if httpLoggerWriter == nil {
		return
	}
	logRequestAndResponseToWriter(httpLoggerWriter, req, rsp, err)
}

func maybeLogFailedResponse(req *http.Request, rsp *http.Response, err error) {
	logsLock()
	defer logsUnlock()

	if !logFailedRequests || isErrLoggingDisabled() {
		return
	}
	if err == nil && rsp.StatusCode < 400 {
		// not failed
		return
	}
	logRequestAndResponseToWriter(httpFailedRequestsLogger, req, rsp, err)
}

// to be able to print request body for failed requests, we must replace
// body with one that captures data read from original body.
func maybeCaptureRequestBody(req *http.Request) {
	shouldCapture := (logFailedRequests && !isErrLoggingDisabled()) || (httpLoggerWriter != nil)
	if !shouldCapture {
		return
	}

	switch req.Method {
	case http.MethodGet, http.MethodHead, "RESET":
		// just in case (probably redundant with req.Bddy != nil check)
		return
	}
	if req.Body != nil {
		req.Body = NewCapturingReadCloser(req.Body)
	}
}

func maybeLogRequestSummary(req *http.Request) {
	if !logRequestSummary {
		return
	}
	method := req.Method
	uri := req.URL.String()
	fmt.Printf("%s %s\n", method, uri)
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
