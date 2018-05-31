package ravendb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
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

func dumpHTTPResponse(resp *http.Response, body []byte) {
	d, err := httputil.DumpResponse(resp, false)
	if err != nil {
		return
	}
	os.Stdout.Write(d)
	if len(body) > 0 {
		os.Stdout.Write(prettyPrintMaybeJSON(body))
		os.Stdout.WriteString("\n")
	}
}

func makeHTTPRequest(n *ServerNode, cmd *RavenCommand) (*http.Request, error) {
	//url := cmd.BuildFullURL(n)
	url := cmd.URLTemplate
	var body io.Reader
	if cmd.Method == http.MethodPut || cmd.Method == http.MethodPost || cmd.Method == http.MethodDelete {
		// TODO: should this be mandatory?
		if cmd.Data != nil {
			body = bytes.NewBuffer(cmd.Data)
		}
	}
	/* TODO:
	   if raven_command.files:
	      data = {"data": data}
	*/
	req, err := http.NewRequest(cmd.Method, url, body)
	if err != nil {
		return nil, err
	}
	for k, v := range cmd.Headers {
		req.Header.Add(k, v)
	}
	req.Header.Add("User-Agent", "ravendb-go-client/1.0")
	req.Header.Add("Raven-Client-Version", "4.0.0.0")
	req.Header.Add("Accept", "application/json")

	// TODO: make logging optional
	fmt.Printf("%s %s\n", cmd.Method, url)

	return req, nil
}

func decodeJSONFromReader(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

// TODO: do I need to explicitly enable compression or does the client does
// it by default? It seems to send Accept-Encoding: gzip by default
func simpleExecutor(n *ServerNode, cmd *RavenCommand) (*http.Response, error) {
	req, err := makeHTTPRequest(n, cmd)
	if err != nil {
		return nil, err
	}

	client := &http.Client{
		Timeout: time.Second * 5,
	}
	rsp, err := client.Do(req)
	// this is for network-level errors when we don't get response
	if err != nil {
		fmt.Printf("client.Do() failed with %s\n", err)
		return nil, err
	}
	// we have response but it could be one of the error server response

	body, _ := getCopyOfResponseBody(rsp)
	dumpHTTPResponse(rsp, body)

	// convert 400 Bad Request response to BadReqeustError
	if rsp.StatusCode == http.StatusBadRequest {
		var res BadRequestError
		err = decodeJSONFromReader(rsp.Body, &res)
		if err != nil {
			return nil, err
		}
		return nil, &res
	}

	// convert 503 Service Unavailable to ServiceUnavailableError
	if rsp.StatusCode == http.StatusServiceUnavailable {
		var res ServiceUnavailableError
		err = decodeJSONFromReader(rsp.Body, &res)
		if err != nil {
			return nil, err
		}
		return nil, &res
	}

	// convert 500 Internal Server to InternalServerError
	if rsp.StatusCode == http.StatusInternalServerError {
		var res InternalServerError
		err = decodeJSONFromReader(rsp.Body, &res)
		if err != nil {
			return nil, err
		}
		return nil, &res
	}

	// convert 409 Conflict to ConflictError
	if rsp.StatusCode == http.StatusConflict {
		var res ConflictError
		err = decodeJSONFromReader(rsp.Body, &res)
		if err != nil {
			return nil, err
		}
		return nil, &res
	}

	// convert 404 Not Found to NotFoundError
	if rsp.StatusCode == http.StatusNotFound {
		// TODO: does it ever return non-empty response?
		res := NotFoundError{
			URL: req.URL.String(),
		}
		return nil, &res
	}

	// TODO: handle other server errors

	isStatusOk := false
	switch rsp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusNoContent:
		isStatusOk = true
	}
	panicIf(!isStatusOk, "unhandled status code %d", rsp.StatusCode)

	return rsp, nil
}
