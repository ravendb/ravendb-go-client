package ravendb

import (
	"fmt"
	"io/ioutil"
	"net/http"
)

type RavenCommand struct {
	data   interface{}
	result interface{}

	statusCode            int
	responseType          RavenCommandResponseType
	_canCache             bool
	_canCacheAggressively bool

	IsReadRequest bool

	// TODO: replace with []*ServerNode
	failedNodes map[*ServerNode]error

	// those simulate Java inheritance in Go
	createRequestFunc  func(c *RavenCommand, node *ServerNode) (*http.Request, error)
	setResponseFunc    func(c *RavenCommand, response string, fromCache bool) error
	setResponseRawFunc func(c *RavenCommand, response *http.Response) error
}

// this is virtual in Java, we set IsReadRequest instead when creating
// RavenCommand instance
func (c *RavenCommand) isReadRequest() bool {
	return c.IsReadRequest
}

func (c *RavenCommand) getResponseType() RavenCommandResponseType {
	return c.responseType
}

func (c *RavenCommand) getStatusCode() int {
	return c.statusCode
}

func (c *RavenCommand) setStatusCode(statusCode int) {
	c.statusCode = statusCode
}

func (c *RavenCommand) getResult() interface{} {
	return c.result
}

func (c *RavenCommand) setResult(result interface{}) {
	c.result = result
}

func (c *RavenCommand) canCache() bool {
	return c._canCache
}

func (c *RavenCommand) canCacheAggressively() bool {
	return c._canCacheAggressively
}

func NewRavenCommand() *RavenCommand {
	res := &RavenCommand{
		responseType:          RavenCommandResponseType_OBJECT,
		_canCache:             true,
		_canCacheAggressively: true,
		createRequestFunc:     defaultCreateRequest,
		setResponseFunc:       defaultSetResponse,
		setResponseRawFunc:    defaultSetResponseRaw,
	}
	return res
}

func defaultSetResponse(c *RavenCommand, response string, fromCache bool) error {
	if c.responseType == RavenCommandResponseType_EMPTY || c.responseType == RavenCommandResponseType_RAW {
		return throwInvalidResponse()
	}

	return NewUnsupportedOperationException(c.responseType + " command must override the setResponse method which expects response with the following type: " + c.responseType)
}

func (c *RavenCommand) setResponse(response String, fromCache bool) error {
	return c.setResponseFunc(c, response, fromCache)
}

func defaultSetResponseRaw(c *RavenCommand, response *http.Response) error {
	panicIf(true, "When "+c.responseType+" is set to Raw then please override this method to handle the response. ")
	return nil
}

func (c *RavenCommand) setResponseRaw(response *http.Response) error {
	return c.setResponseRawFunc(c, response)
}

func defaultCreateRequest(c *RavenCommand, node *ServerNode) (*http.Request, error) {
	panicIf(true, "must over-write createRequestFunc")
	return nil, nil
}

func (c *RavenCommand) createRequest(node *ServerNode) (*http.Request, error) {
	return c.createRequestFunc(c, node)
}

func throwInvalidResponse() error {
	return fmt.Errorf("Invalid response")
}

func (c *RavenCommand) send(client *http.Client, request *http.Request) (*http.Response, error) {
	return client.Do(request)
}

func (c *RavenCommand) getFailedNodes() map[*ServerNode]error {
	return c.failedNodes
}

func (c *RavenCommand) setFailedNodes(failedNodes map[*ServerNode]error) {
	c.failedNodes = failedNodes
}

func (c *RavenCommand) urlEncode(value String) string {
	return urlEncode(value)
}

// TODO: return error?
func ensureIsNotNullOrString(value String, name String) {
	panicIf(value == "", "%s", name+" cannot be null or empty")
}

func (c *RavenCommand) isFailedWithNode(node *ServerNode) bool {
	if c.failedNodes == nil {
		return false
	}
	_, ok := c.failedNodes[node]
	return ok
}

func (c *RavenCommand) processResponse(cache *HttpCache, response *http.Response, url String) (ResponseDisposeHandling, error) {
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
		err = c.setResponse(string(js), false)
		return ResponseDisposeHandling_AUTOMATIC, err
	} else {
		c.setResponseRaw(response)
	}

	return ResponseDisposeHandling_AUTOMATIC, nil
}

func (c *RavenCommand) cacheResponse(cache *HttpCache, url String, response *http.Response, responseJson String) {
	if !c.canCache() {
		return
	}

	changeVector := HttpExtensions_getEtagHeader(response)
	if changeVector == nil {
		return
	}

	cache.set(url, *changeVector, responseJson)
}

func (c *RavenCommand) addChangeVectorIfNotNull(changeVector *String, request *http.Request) {
	if changeVector != nil {
		request.Header.Add("If-Match", `"`+*changeVector+`"`)
	}
}

func (c *RavenCommand) onResponseFailure(response *http.Response) {
	// TODO: it looks like it's meant to be virtual but there are no
	// over-rides in Java code
}
