package ravendb

import (
	"io"
	"io/ioutil"
	"net/http"
)

var (
	_ RavenCommand = &RavenCommandBase{}
)

// RavenCommand defines interface for server commands
type RavenCommand interface {
	// those are meant to be over-written
	CreateRequest(node *ServerNode) (*http.Request, error)
	SetResponse(response []byte, fromCache bool) error
	SetResponseRaw(response *http.Response, body io.Reader) error

	Send(client *http.Client, req *http.Request) (*http.Response, error)

	// for all other functions, get access to underlying RavenCommandBase
	GetBase() *RavenCommandBase
}

type RavenCommandBase struct {
	StatusCode           int
	ResponseType         RavenCommandResponseType
	CanCache             bool
	CanCacheAggressively bool

	// if true, can be cached
	IsReadRequest bool

	FailedNodes map[*ServerNode]error
}

func NewRavenCommandBase() RavenCommandBase {
	res := RavenCommandBase{
		ResponseType:         RavenCommandResponseTypeObject,
		CanCache:             true,
		CanCacheAggressively: true,
	}
	return res
}

func (c *RavenCommandBase) GetBase() *RavenCommandBase {
	return c
}

func (c *RavenCommandBase) SetResponse(response []byte, fromCache bool) error {
	if c.ResponseType == RavenCommandResponseTypeEmpty || c.ResponseType == RavenCommandResponseTypeRaw {
		return throwInvalidResponse()
	}

	return newUnsupportedOperationError(c.ResponseType + " command must override the SetResponse method which expects response with the following type: " + c.ResponseType)
}

func (c *RavenCommandBase) SetResponseRaw(response *http.Response, stream io.Reader) error {
	panicIf(true, "When "+c.ResponseType+" is set to Raw then please override this method to handle the response. ")
	return nil
}

func (c *RavenCommandBase) CreateRequest(node *ServerNode) (*http.Request, error) {
	panicIf(true, "CreateRequest must be over-written by all types")
	return nil, nil
}

func throwInvalidResponse() error {
	return newIllegalStateError("Invalid response")
}

func (c *RavenCommandBase) Send(client *http.Client, req *http.Request) (*http.Response, error) {
	rsp, err := client.Do(req)
	return rsp, err
}

func (c *RavenCommandBase) urlEncode(value string) string {
	return urlEncode(value)
}

func ensureIsNotNullOrString(value string, name string) error {
	if value == "" {
		return newIllegalArgumentError("%s cannot be null or empty", name)
	}
	return nil
}

// Note: unused
func (c *RavenCommandBase) isFailedWithNode(node *ServerNode) bool {
	if c.FailedNodes == nil {
		return false
	}
	_, ok := c.FailedNodes[node]
	return ok
}

// Note: in Java Raven.processResponse is virtual.
// That's impossible in Go, so we replace with stand-alone function that dispatches based on type
func ravenCommand_processResponse(cmd RavenCommand, cache *httpCache, response *http.Response, url string) (responseDisposeHandling, error) {
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
		return responseDisposeHandlingAutomatic, nil
	}

	statusCode := response.StatusCode
	if c.ResponseType == RavenCommandResponseTypeEmpty || statusCode == http.StatusNoContent {
		return responseDisposeHandlingAutomatic, nil
	}

	if c.ResponseType == RavenCommandResponseTypeObject {
		contentLength := response.ContentLength
		if contentLength == 0 {
			return responseDisposeHandlingAutomatic, nil
		}

		// we intentionally don't dispose the reader here, we'll be using it
		// in the command, any associated memory will be released on context reset
		js, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return responseDisposeHandlingAutomatic, err
		}

		if cache != nil {
			c.cacheResponse(cache, url, response, js)
		}
		err = cmd.SetResponse(js, false)
		return responseDisposeHandlingAutomatic, err
	}

	err := cmd.SetResponseRaw(response, response.Body)
	return responseDisposeHandlingAutomatic, err
}

func (c *RavenCommandBase) cacheResponse(cache *httpCache, url string, response *http.Response, responseJson []byte) {
	if !c.CanCache {
		return
	}

	changeVector := gttpExtensionsGetEtagHeader(response)
	if changeVector == nil {
		return
	}

	cache.set(url, changeVector, responseJson)
}

// Note: unused
func (c *RavenCommandBase) addChangeVectorIfNotNull(changeVector *string, request *http.Request) {
	if changeVector != nil {
		request.Header.Add("If-Match", `"`+*changeVector+`"`)
	}
}

func (c *RavenCommandBase) onResponseFailure(response *http.Response) {
	// Note: it looks like it's meant to be virtual but there are no
	// over-rides in Java code
}

// Note: hackish solution due to lack of generics
// Returns OperationIDReuslt for commands that have it as a result
// When new command returning OperationIDResult are added, we must extend it
func getCommandOperationIDResult(cmd RavenCommand) *OperationIDResult {
	switch c := cmd.(type) {
	case *CompactDatabaseCommand:
		return c.Result
	case *PatchByQueryCommand:
		return c.Result
	case *DeleteByIndexCommand:
		return c.Result
	}

	panicIf(true, "called on a command %T that doesn't return OperationIDResult", cmd)
	return nil
}
