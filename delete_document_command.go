package ravendb

import "net/http"

var (
	_ RavenCommand = &DeleteDocumentCommand{}
)

type DeleteDocumentCommand struct {
	*RavenCommandBase
	_id           String
	_changeVector String
}

func NewDeleteDocumentCommand(id String, changeVector String) *DeleteDocumentCommand {
	cmd := &DeleteDocumentCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_id:           id,
		_changeVector: changeVector,
	}
	cmd.responseType = RavenCommandResponseType_EMPTY
	return cmd
}

func (c *DeleteDocumentCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/docs?id=" + urlEncode(c._id)

	request, err := NewHttpDelete(url, "")
	if err != nil {
		return nil, err
	}
	addChangeVectorIfNotNull(c._changeVector, request)
	return request, nil

}
