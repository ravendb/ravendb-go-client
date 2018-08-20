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
	cmd.ResponseType = RavenCommandResponseType_EMPTY
	return cmd
}

func (c *DeleteDocumentCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/docs?id=" + urlEncode(c._id)

	request, err := NewHttpDelete(url, nil)
	if err != nil {
		return nil, err
	}
	addChangeVectorIfNotNull(c._changeVector, request)
	return request, nil

}
