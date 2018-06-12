package ravendb

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

var (
	_ RavenCommand = &RavenCommandBase{}
)

type RavenCommand interface {
	// those are meant to be over-written
	createRequest(node *ServerNode) (*http.Request, error)
	setResponse(response string, fromCache bool) error
	setResponseRaw(response *http.Response, body io.Reader) error

	// for all other functions, get access to underlying RavenCommandBase
	getBase() *RavenCommandBase
}

type RavenCommandBase struct {
	statusCode            int
	responseType          RavenCommandResponseType
	_canCache             bool
	_canCacheAggressively bool

	IsReadRequest bool

	failedNodes map[*ServerNode]error
}

func NewRavenCommandBase() *RavenCommandBase {
	res := &RavenCommandBase{
		responseType:          RavenCommandResponseType_OBJECT,
		_canCache:             true,
		_canCacheAggressively: true,
	}
	return res
}

func (c *RavenCommandBase) getBase() *RavenCommandBase {
	return c
}

// this is virtual in Java, we set IsReadRequest instead when creating
// RavenCommand instance
func (c *RavenCommandBase) isReadRequest() bool {
	return c.IsReadRequest
}

func (c *RavenCommandBase) getResponseType() RavenCommandResponseType {
	return c.responseType
}

func (c *RavenCommandBase) getStatusCode() int {
	return c.statusCode
}

func (c *RavenCommandBase) setStatusCode(statusCode int) {
	c.statusCode = statusCode
}

func (c *RavenCommandBase) canCache() bool {
	return c._canCache
}

func (c *RavenCommandBase) canCacheAggressively() bool {
	return c._canCacheAggressively
}

func (c *RavenCommandBase) setResponse(response String, fromCache bool) error {
	if c.responseType == RavenCommandResponseType_EMPTY || c.responseType == RavenCommandResponseType_RAW {
		return throwInvalidResponse()
	}

	return NewUnsupportedOperationException(c.responseType + " command must override the setResponse method which expects response with the following type: " + c.responseType)
}

// TODO: this is only implemented on MultiGetCommand
func (c *RavenCommandBase) setResponseRaw(response *http.Response, stream io.Reader) error {
	panicIf(true, "When "+c.responseType+" is set to Raw then please override this method to handle the response. ")
	return nil
}

func (c *RavenCommandBase) createRequest(node *ServerNode) (*http.Request, error) {
	panicIf(true, "must over-write createRequestFunc")
	return nil, nil
}

func throwInvalidResponse() error {
	return fmt.Errorf("Invalid response")
}

func (c *RavenCommandBase) send(client *http.Client, request *http.Request) (*http.Response, error) {
	return client.Do(request)
}

func (c *RavenCommandBase) getFailedNodes() map[*ServerNode]error {
	return c.failedNodes
}

func (c *RavenCommandBase) setFailedNodes(failedNodes map[*ServerNode]error) {
	c.failedNodes = failedNodes
}

func (c *RavenCommandBase) urlEncode(value String) string {
	return urlEncode(value)
}

// TODO: return error?
func ensureIsNotNullOrString(value String, name String) {
	panicIf(value == "", "%s", name+" cannot be null or empty")
}

func (c *RavenCommandBase) isFailedWithNode(node *ServerNode) bool {
	if c.failedNodes == nil {
		return false
	}
	_, ok := c.failedNodes[node]
	return ok
}

// Note: in Java this is part of RavenCommand and can be virtual
// That's imposssible in Go, so we replace with stand-alone function
func processCommandResponse(cmd RavenCommand, cache *HttpCache, response *http.Response, url String) (ResponseDisposeHandling, error) {
	// In Java this is overridden in HeadDocumentCommand, so hack it this way
	if cmdHead, ok := cmd.(*HeadDocumentCommand); ok {
		return cmdHead.processResponse(cache, response, url)
	}

	c := cmd.getBase()

	if response.Body == nil {
		return ResponseDisposeHandling_AUTOMATIC, nil
	}

	statusCode := response.StatusCode
	if c.responseType == RavenCommandResponseType_EMPTY || statusCode == http.StatusNoContent {
		return ResponseDisposeHandling_AUTOMATIC, nil
	}

	if c.responseType == RavenCommandResponseType_OBJECT {
		contentLength := response.ContentLength
		if contentLength == 0 {
			return ResponseDisposeHandling_AUTOMATIC, nil
		}

		// we intentionally don't dispose the reader here, we'll be using it
		// in the command, any associated memory will be released on context reset
		js, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return ResponseDisposeHandling_AUTOMATIC, err
		}
		if cache != nil {
			c.cacheResponse(cache, url, response, string(js))
		}
		err = cmd.setResponse(string(js), false)
		return ResponseDisposeHandling_AUTOMATIC, err
	} else {
		cmd.setResponseRaw(response, response.Body)
	}

	return ResponseDisposeHandling_AUTOMATIC, nil
}

func (c *RavenCommandBase) cacheResponse(cache *HttpCache, url String, response *http.Response, responseJson String) {
	if !c.canCache() {
		return
	}

	changeVector := HttpExtensions_getEtagHeader(response)
	if changeVector == nil {
		return
	}

	cache.set(url, *changeVector, responseJson)
}

func (c *RavenCommandBase) addChangeVectorIfNotNull(changeVector *String, request *http.Request) {
	if changeVector != nil {
		request.Header.Add("If-Match", `"`+*changeVector+`"`)
	}
}

func (c *RavenCommandBase) onResponseFailure(response *http.Response) {
	// TODO: it looks like it's meant to be virtual but there are no
	// over-rides in Java code
}
