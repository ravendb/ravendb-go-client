package ravendb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
)

// BufferCloser is a wrapper around bytes.Buffer that adds io.Close method
// to make it io.ReadCloser
type BufferCloser struct {
	*bytes.Buffer
}

// NewBufferCloser creates new BufferClose
func NewBufferCloser(buf *bytes.Buffer) *BufferCloser {
	if buf == nil {
		buf = &bytes.Buffer{}
	}
	return &BufferCloser{
		Buffer: buf,
	}
}

// Close implements io.Close interface
func (b *BufferCloser) Close() error {
	// nothing to do
	return nil
}

// retruns copy of resp.Body but also makes it available for subsequent reads
func getCopyOfResponseBody(resp *http.Response) ([]byte, error) {
	if resp == nil {
		return nil, nil
	}
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	resp.Body = NewBufferCloser(bytes.NewBuffer(d))
	return d, nil
}

// if d is a valid json, pretty-print it
func prettyPrintMaybeJSON(d []byte) []byte {
	var m map[string]interface{}
	err := json.Unmarshal(d, &m)
	if err != nil {
		return d
	}
	d2, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return d
	}
	return d2
}

var (
	dumpHTTP     bool
	dumpHTTPBody bool
)

// TODO: also dump body
func dumpHTTPRequest(req *http.Request) {
	if !dumpHTTP {
		return
	}
	d, err := httputil.DumpRequest(req, false)
	if err != nil {
		fmt.Printf("httputil.DumpRequest failed with %s\n", err)
		return
	}
	io.WriteString(os.Stdout, "HTTP REQUEST:\n")
	os.Stdout.Write(d)
}

func dumpHTTPResponse(resp *http.Response) {
	if !dumpHTTP {
		return
	}
	d, err := httputil.DumpResponse(resp, false)
	if err != nil {
		fmt.Printf("httputil.DumpResponse failed with %s\n", err)
		return
	}
	io.WriteString(os.Stdout, "HTTP RESPONSE:\n")
	os.Stdout.Write(d)

	if !dumpHTTPBody {
		return
	}
	body, err := getCopyOfResponseBody(resp)
	if err != nil {
		os.Stdout.Write(prettyPrintMaybeJSON(body))
		os.Stdout.WriteString("\n")
	}
}

func dumpHTTPRequestAndResponse(req *http.Request, resp *http.Response) {
	dumpHTTPRequest(req)
	dumpHTTPResponse(resp)
}

func decodeJSONFromReader(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
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

func NewHttpHead(uri string) (*http.Request, error) {
	//fmt.Printf("GET %s\n", uri)
	req, err := http.NewRequest(http.MethodHead, uri, nil)
	if err != nil {
		return nil, err
	}
	addCommonHeaders(req)
	return req, err
}

func NewHttpGet(uri string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	addCommonHeaders(req)
	return req, err
}

func NewHttpReset(uri string) (*http.Request, error) {
	req, err := http.NewRequest("RESET", uri, nil)
	if err != nil {
		return nil, err
	}
	addCommonHeaders(req)
	return req, err
}

func NewHttpPost(uri string, data []byte) (*http.Request, error) {
	var body io.Reader
	if len(data) > 0 {
		body = bytes.NewBuffer(data)
		//d := prettyPrintMaybeJSON([]byte(data))
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
	return req, err
}

func NewHttpPut(uri string, data []byte) (*http.Request, error) {
	var body io.Reader
	if len(data) > 0 {
		body = bytes.NewBuffer(data)
		//d := prettyPrintMaybeJSON([]byte(data))
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
	return req, err
}

func NewHttpPatch(uri string, data []byte) (*http.Request, error) {
	var body io.Reader
	if len(data) > 0 {
		body = bytes.NewBuffer(data)
		//d := prettyPrintMaybeJSON([]byte(data))
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
	return req, err
}

func NewHttpDelete(uri string, data []byte) (*http.Request, error) {
	var body io.Reader
	if len(data) > 0 {
		body = bytes.NewBuffer(data)
		//d := prettyPrintMaybeJSON([]byte(data))
		//fmt.Printf("%s\n", string(d))
	}
	req, err := http.NewRequest(http.MethodDelete, uri, body)
	if err != nil {
		return nil, err
	}
	addCommonHeaders(req)
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	return req, nil
}
