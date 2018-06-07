package ravendb

import "net/http"

type _DeleteDocumentCommand struct {
	_id           String
	_changeVector String
}

func NewDeleteDocumentCommand(id String, changeVector String) *RavenCommand {
	data := &_DeleteDocumentCommand{
		_id:           id,
		_changeVector: changeVector,
	}
	cmd := NewRavenCommand()
	cmd.data = data
	cmd.createRequestFunc = DeleteDocumentCommand_createRequest
	return cmd
}

func DeleteDocumentCommand_createRequest(cmd *RavenCommand, node *ServerNode) (*http.Request, error) {
	data := cmd.data.(*_DeleteDocumentCommand)
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/docs?id=" + urlEncode(data._id)

	request, err := NewHttpDelete(url, "")
	if err != nil {
		return nil, err
	}
	addChangeVectorIfNotNull(data._changeVector, request)
	return request, nil

}