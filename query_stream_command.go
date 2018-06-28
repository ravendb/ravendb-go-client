package ravendb

import (
	"encoding/json"
	"net/http"
)

var (
	_ RavenCommand = &QueryStreamCommand{}
)

type QueryStreamCommand struct {
	*RavenCommandBase

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

func (c *QueryStreamCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/streams/queries"

	m := JsonExtensions_writeIndexQuery(c._conventions, c._indexQuery)
	d, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return NewHttpPost(url, d)
}

func (c *QueryStreamCommand) processResponse(cache *HttpCache, response *http.Response, url string) (ResponseDisposeHandling, error) {

	// TODO: return an error if response.Body is nil
	streamResponse := NewStreamResultResponse()
	streamResponse.setResponse(response)
	streamResponse.setStream(response.Body)
	c.Result = streamResponse

	return ResponseDisposeHandling_MANUALLY, nil
}
