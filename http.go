package ravendb

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

var ()

func urlEncode(s string) string {
	return url.PathEscape(s)
}

func addChangeVectorIfNotNull(changeVector *string, req *http.Request) {
	if changeVector != nil {
		req.Header.Add("If-Match", fmt.Sprintf(`"%s"`, *changeVector))
	}
}

func addCommonHeaders(req *http.Request) {
	req.Header.Add("User-Agent", "ravendb-go-client/4.0.0")
}

func NewHttpHead(uri string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodHead, uri, nil)
	if err != nil {
		return nil, err
	}
	addCommonHeaders(req)
	return req, nil
}

func NewHttpGet(uri string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, err
	}
	addCommonHeaders(req)
	return req, nil
}

func NewHttpReset(uri string) (*http.Request, error) {
	req, err := http.NewRequest("RESET", uri, nil)
	if err != nil {
		return nil, err
	}
	addCommonHeaders(req)
	return req, nil
}

func NewHttpPostReader(uri string, r io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, uri, r)
	if err != nil {
		return nil, err
	}
	addCommonHeaders(req)
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
	return req, nil
}

func NewHttpPutReader(uri string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPut, uri, body)
	if err != nil {
		return nil, err
	}
	addCommonHeaders(req)
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
	return req, nil
}
