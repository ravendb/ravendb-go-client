package ravendb

import "net/http"

var (
	_ RavenCommand = &DeleteDocumentCommand{}
)

type DeleteDocumentCommand struct {
	RavenCommandBase

	_id           string
	_changeVector *string
}

func NewDeleteDocumentCommand(id string, changeVector *string) *DeleteDocumentCommand {
	cmd := &DeleteDocumentCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_id:           id,
		_changeVector: changeVector,
	}
	cmd.ResponseType = RavenCommandResponseTypeEmpty
	return cmd
}

func (c *DeleteDocumentCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/docs?id=" + urlEncode(c._id)

	request, err := newHttpDelete(url, nil)
	if err != nil {
		return nil, err
	}
	addChangeVectorIfNotNull(c._changeVector, request)
	return request, nil

}
