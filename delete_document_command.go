package ravendb

import "net/http"

type DeleteDocumentCommandData struct {
	_id           String
	_changeVector String
}

func NewDeleteDocumentCommand(id String, changeVector String) *RavenCommand {
	data := &DeleteDocumentCommandData{
		_id:           id,
		_changeVector: changeVector,
	}
	cmd := NewRavenCommand()
	cmd.data = data
	cmd.createRequestFunc = DeleteDocumentCommand_createRequest
	//cmd.setResponseFunc = DeleteDocumentCommand_setResponse
	return cmd
}

func DeleteDocumentCommand_createRequest(cmd *RavenCommand, node *ServerNode) (*http.Request, string) {
	data := cmd.data.(*DeleteDocumentCommandData)
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/docs?id=" + urlEncode(data._id)

	request := NewHttpDelete(url, "")
	addChangeVectorIfNotNull(data._changeVector, request)
	return request, url

}
