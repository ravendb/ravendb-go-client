package ravendb

import (
	"net/http"
)

var (
	_ RavenCommand = &QueryStreamCommand{}
)

type QueryStreamCommand struct {
	RavenCommandBase

	_conventions *DocumentConventions
	_indexQuery  *IndexQuery

	Result *StreamResultResponse
}

func NewQueryStreamCommand(conventions *DocumentConventions, indexQuery *IndexQuery) *QueryStreamCommand {
	panicIf(indexQuery == nil, "IndexQuery cannot be null")
	// TODO: validate convention
	cmd := &QueryStreamCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_conventions: conventions,
		_indexQuery:  indexQuery,
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *QueryStreamCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/streams/queries"

	m := jsonExtensionsWriteIndexQuery(c._conventions, c._indexQuery)
	d, err := jsonMarshal(m)
	if err != nil {
		return nil, err
	}
	return NewHttpPost(url, d)
}

func (c *QueryStreamCommand) processResponse(cache *HttpCache, response *http.Response, url string) (responseDisposeHandling, error) {

	// TODO: return an error if response.Body is nil
	streamResponse := &StreamResultResponse{
		Response: response,
		Stream:   response.Body,
	}
	c.Result = streamResponse

	return responseDisposeHandlingManually, nil
}
