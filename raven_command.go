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
	CreateRequest(node *ServerNode) (*http.Request, error)
	SetResponse(response []byte, fromCache bool) error
	SetResponseRaw(response *http.Response, body io.Reader) error

	// for all other functions, get access to underlying RavenCommandBase
	GetBase() *RavenCommandBase
}

type RavenCommandBase struct {
	statusCode            int
	ResponseType          RavenCommandResponseType
	_canCache             bool
	_canCacheAggressively bool

	// if true, can be cached
	IsReadRequest bool

	failedNodes map[*ServerNode]error
}

func NewRavenCommandBase() *RavenCommandBase {
	res := &RavenCommandBase{
		ResponseType:          RavenCommandResponseType_OBJECT,
		_canCache:             true,
		_canCacheAggressively: true,
	}
	return res
}

func (c *RavenCommandBase) GetBase() *RavenCommandBase {
	return c
}

// this is virtual in Java, we set IsReadRequest instead when creating
// RavenCommand instance
func (c *RavenCommandBase) isReadRequest() bool {
	return c.IsReadRequest
}

func (c *RavenCommandBase) GetResponseType() RavenCommandResponseType {
	return c.ResponseType
}

func (c *RavenCommandBase) GetStatusCode() int {
	return c.statusCode
}

func (c *RavenCommandBase) SetStatusCode(statusCode int) {
	c.statusCode = statusCode
}

func (c *RavenCommandBase) CanCache() bool {
	return c._canCache
}

func (c *RavenCommandBase) CanCacheAggressively() bool {
	return c._canCacheAggressively
}

func (c *RavenCommandBase) SetResponse(response []byte, fromCache bool) error {
	if c.ResponseType == RavenCommandResponseType_EMPTY || c.ResponseType == RavenCommandResponseType_RAW {
		return throwInvalidResponse()
	}

	return NewUnsupportedOperationException(c.ResponseType + " command must override the SetResponse method which expects response with the following type: " + c.ResponseType)
}

// TODO: this is only implemented on MultiGetCommand
func (c *RavenCommandBase) SetResponseRaw(response *http.Response, stream io.Reader) error {
	panicIf(true, "When "+c.ResponseType+" is set to Raw then please override this method to handle the response. ")
	return nil
}

func (c *RavenCommandBase) CreateRequest(node *ServerNode) (*http.Request, error) {
	panicIf(true, "must over-write createRequestFunc")
	return nil, nil
}

func throwInvalidResponse() error {
	return fmt.Errorf("Invalid response")
}

func (c *RavenCommandBase) Send(client *http.Client, req *http.Request) (*http.Response, error) {
	HTTPRequestCount.IncrementAndGet()
	rsp, err := client.Do(req)
	maybeLogFailedResponse(req, rsp, err)
	maybeLogHTTPRequest(req, rsp, err)
	return rsp, err
}

func (c *RavenCommandBase) GetFailedNodes() map[*ServerNode]error {
	return c.failedNodes
}

func (c *RavenCommandBase) SetFailedNodes(failedNodes map[*ServerNode]error) {
	c.failedNodes = failedNodes
}

func (c *RavenCommandBase) urlEncode(value string) string {
	return urlEncode(value)
}

func ensureIsNotNullOrString(value string, name string) error {
	if value == "" {
		return NewIllegalArgumentException("%s cannot be null or empty", name)
	}
	return nil
}

func (c *RavenCommandBase) IsFailedWithNode(node *ServerNode) bool {
	if c.failedNodes == nil {
		return false
	}
	_, ok := c.failedNodes[node]
	return ok
}

// Note: in Java this is part of RavenCommand and can be virtual
// That's imposssible in Go, so we replace with stand-alone function
func processCommandResponse(cmd RavenCommand, cache *HttpCache, response *http.Response, url string) (ResponseDisposeHandling, error) {
	// In Java this is overridden in HeadDocumentCommand, so hack it this way
	if cmdHead, ok := cmd.(*HeadDocumentCommand); ok {
		return cmdHead.ProcessResponse(cache, response, url)
	}

	if cmdHead, ok := cmd.(*HeadAttachmentCommand); ok {
		return cmdHead.processResponse(cache, response, url)
	}

	if cmdGet, ok := cmd.(*GetAttachmentCommand); ok {
		return cmdGet.processResponse(cache, response, url)
	}

	if cmdQuery, ok := cmd.(*QueryStreamCommand); ok {
		return cmdQuery.processResponse(cache, response, url)
	}

	if cmdStream, ok := cmd.(*StreamCommand); ok {
		return cmdStream.processResponse(cache, response, url)
	}

	c := cmd.GetBase()

	if response.Body == nil {
		return ResponseDisposeHandling_AUTOMATIC, nil
	}

	statusCode := response.StatusCode
	if c.ResponseType == RavenCommandResponseType_EMPTY || statusCode == http.StatusNoContent {
		return ResponseDisposeHandling_AUTOMATIC, nil
	}

	if c.ResponseType == RavenCommandResponseType_OBJECT {
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
			c.CacheResponse(cache, url, response, string(js))
		}
		err = cmd.SetResponse(js, false)
		return ResponseDisposeHandling_AUTOMATIC, err
	} else {
		cmd.SetResponseRaw(response, response.Body)
	}

	return ResponseDisposeHandling_AUTOMATIC, nil
}

func (c *RavenCommandBase) CacheResponse(cache *HttpCache, url string, response *http.Response, responseJson string) {
	if !c.CanCache() {
		return
	}

	changeVector := HttpExtensions_getEtagHeader(response)
	if changeVector == nil {
		return
	}

	cache.set(url, *changeVector, responseJson)
}

func (c *RavenCommandBase) AddChangeVectorIfNotNull(changeVector *string, request *http.Request) {
	if changeVector != nil {
		request.Header.Add("If-Match", `"`+*changeVector+`"`)
	}
}

func (c *RavenCommandBase) OnResponseFailure(response *http.Response) {
	// TODO: it looks like it's meant to be virtual but there are no
	// over-rides in Java code
}

// Note: hackish solution due to lack of generics
// For commands whose result is OperationIdResult, return it
// When new command returning OperationIdResult are added, we must extend it
func getCommandOperationIdResult(cmd RavenCommand) *OperationIdResult {
	switch c := cmd.(type) {
	case *CompactDatabaseCommand:
		return c.Result
	case *PatchByQueryCommand:
		return c.Result
	case *DeleteByIndexCommand:
		return c.Result
	}

	panicIf(true, "called on a command %T that doesn't return OperationIdResult", cmd)
	return nil
}
