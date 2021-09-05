package ravendb

import (
	"net/http"
)

var (
	_ RavenCommand = &StreamCommand{}
)

type StreamCommand struct {
	RavenCommandBase

	_url string

	Result *StreamResultResponse
}

func NewStreamCommand(url string) *StreamCommand {
	// TODO: validate url
	cmd := &StreamCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_url: url,
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *StreamCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/" + c._url
	return newHttpGet(url)
}

func (c *StreamCommand) processResponse(cache *httpCache, response *http.Response, url string) (responseDisposeHandling, error) {

	// TODO: return an error if response.Body is nil
	streamResponse := &StreamResultResponse{
		Response: response,
		Stream:   response.Body,
	}
	c.Result = streamResponse

	return responseDisposeHandlingManually, nil
}
