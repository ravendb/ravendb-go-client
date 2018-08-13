package ravendb

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

var (
	// if true, prints requests and their responses to stdout
	gLogHTTP = false
	// if not nil, we write all http requests and responses here
	gHTTPLogger io.WriteCloser
	// numbering of requests helps match http traffic from java client with go client
	gHTTPRequestCount AtomicInteger
)

// retruns copy of resp.Body but also makes it available for subsequent reads
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

func dumpRequestAndResponseToWriter(w io.Writer, req *http.Request, rsp *http.Response, reqErr error) {
	n := gHTTPRequestCount.get()

	fmt.Fprintf(w, "=========== %d:\n", n)
	if reqErr != nil {
		fmt.Fprintf(w, "%s %s failed with '%s'\n", req.Method, req.URL, reqErr)
	}

	d, err := httputil.DumpRequest(req, false)
	if err == nil {
		w.Write(d)
	}
	if reqErr != nil {
		return
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

func maybeLogHTTPRequest(req *http.Request, rsp *http.Response, err error) {
	if gHTTPLogger == nil {
		return
	}
	dumpRequestAndResponseToWriter(gHTTPLogger, req, rsp, err)
}

func dumpRequestAndResponse(req *http.Request, rsp *http.Response, err error) {
	dumpRequestAndResponseToWriter(os.Stdout, req, rsp, err)
}

func maybeDumpFailedResponse(req *http.Request, rsp *http.Response, err error) {
	if !gLogFailedRequests {
		return
	}
	if err == nil && rsp.StatusCode < 400 {
		// not failed
		return
	}
	dumpRequestAndResponse(req, rsp, err)
}

func urlEncode(s string) string {
	return url.PathEscape(s)
}

func addChangeVectorIfNotNull(changeVector *string, req *http.Request) {
	if changeVector != nil {
		req.Header.Add("If-Match", fmt.Sprintf(`"%s"`, *changeVector))
	}
}

func addCommonHeaders(req *http.Request) {
	req.Header.Add("User-Agent", "ravendb-go-client/1.0")
}

// to be able to print request body for failed requests, we must replace
// body with one that captures data read from original body.
func maybeCaptureRequestBody(req *http.Request) {
	shouldCapture := gLogFailedRequests || gLogHTTP || (gHTTPLogger != nil)
	if !shouldCapture {
		return
	}
	if req.Body != nil {
		req.Body = NewCapturingReadCloser(req.Body)
	}
}

func maybeLogRequestSummary(req *http.Request) {
	if !gLogRequestSummary {
		return
	}
	method := req.Method
	uri := req.URL.String()
	fmt.Printf("%s %s\n", method, uri)
}

func NewHttpHead(uri string) (*http.Request, error) {
	//fmt.Printf("GET %s\n", uri)
	req, err := http.NewRequest(http.MethodHead, uri, nil)
	if err != nil {
		return nil, err
	}
	addCommonHeaders(req)
	maybeLogRequestSummary(req)
	return req, nil
}

func NewHttpGet(uri string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	addCommonHeaders(req)
	maybeLogRequestSummary(req)
	return req, nil
}

func NewHttpReset(uri string) (*http.Request, error) {
	req, err := http.NewRequest("RESET", uri, nil)
	if err != nil {
		return nil, err
	}
	addCommonHeaders(req)
	maybeLogRequestSummary(req)
	return req, nil
}

func NewHttpPostReader(uri string, r io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, uri, r)
	if err != nil {
		return nil, err
	}
	addCommonHeaders(req)
	maybeLogRequestSummary(req)
	maybeCaptureRequestBody(req)
	return req, nil
}

func NewHttpPost(uri string, data []byte) (*http.Request, error) {
	var body io.Reader
	if len(data) > 0 {
		body = bytes.NewReader(data)
		//d := maybePrettyPrintJSON([]byte(data))
		//fmt.Printf("%s\n", string(d))
	}
	req, err := http.NewRequest(http.MethodPost, uri, body)
	if err != nil {
		return nil, err
	}
	addCommonHeaders(req)
	if len(data) > 0 {
		req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	}
	maybeLogRequestSummary(req)
	maybeCaptureRequestBody(req)
	return req, nil
}

func NewHttpPut(uri string, data []byte) (*http.Request, error) {
	var body io.Reader
	if len(data) > 0 {
		body = bytes.NewReader(data)
		//d := maybePrettyPrintJSON([]byte(data))
		//fmt.Printf("%s\n", string(d))
	}
	req, err := http.NewRequest(http.MethodPut, uri, body)
	if err != nil {
		return nil, err
	}
	addCommonHeaders(req)
	if len(data) > 0 {
		req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	}
	maybeLogRequestSummary(req)
	maybeCaptureRequestBody(req)
	return req, nil
}

func NewHttpPutReader(uri string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPut, uri, body)
	if err != nil {
		return nil, err
	}
	addCommonHeaders(req)
	maybeLogRequestSummary(req)
	maybeCaptureRequestBody(req)
	return req, nil
}

func NewHttpPatch(uri string, data []byte) (*http.Request, error) {
	var body io.Reader
	if len(data) > 0 {
		body = bytes.NewReader(data)
		//d := maybePrettyPrintJSON([]byte(data))
		//fmt.Printf("%s\n", string(d))
	}
	req, err := http.NewRequest(http.MethodPatch, uri, body)
	if err != nil {
		return nil, err
	}
	addCommonHeaders(req)
	if len(data) > 0 {
		req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	}
	maybeLogRequestSummary(req)
	maybeCaptureRequestBody(req)
	return req, nil
}

func NewHttpDelete(uri string, data []byte) (*http.Request, error) {
	var body io.Reader
	if len(data) > 0 {
		body = bytes.NewReader(data)
		//d := maybePrettyPrintJSON([]byte(data))
		//fmt.Printf("%s\n", string(d))
	}
	req, err := http.NewRequest(http.MethodDelete, uri, body)
	if err != nil {
		return nil, err
	}
	addCommonHeaders(req)
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	maybeLogRequestSummary(req)
	maybeCaptureRequestBody(req)
	return req, nil
}
