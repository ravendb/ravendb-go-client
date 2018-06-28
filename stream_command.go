package ravendb

import (
	"net/http"
)

var (
	_ RavenCommand = &StreamCommand{}
)

type StreamCommand struct {
	*RavenCommandBase

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

func (c *StreamCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/" + c._url
	return NewHttpGet(url)
}

func (c *StreamCommand) processResponse(cache *HttpCache, response *http.Response, url string) (ResponseDisposeHandling, error) {

	// TODO: return an error if response.Body is nil
	streamResponse := NewStreamResultResponse()
	streamResponse.setResponse(response)
	streamResponse.setStream(response.Body)
	c.Result = streamResponse

	return ResponseDisposeHandling_MANUALLY, nil
}
