package ravendb

import (
	"encoding/json"
	"net/http"
)

type _HeadDocumentCommand struct {
	_id           string
	_changeVector *string
}

func NewHeadDocumentCommand(id string, changeVector *string) *RavenCommand {
	panicIf(id == "", "id cannot be empty")
	data := &_HeadDocumentCommand{
		_id:           id,
		_changeVector: changeVector,
	}

	cmd := NewRavenCommand()
	cmd.data = data
	cmd.createRequestFunc = HeadDocumentCommand_createRequest
	cmd.setResponseFunc = HeadDocumentCommand_setResponse

	return cmd
}

func HeadDocumentCommand_createRequest(cmd *RavenCommand, node *ServerNode) (*http.Request, error) {
	data := cmd.data.(*_HeadDocumentCommand)

	url := node.getUrl() + "/databases/" + node.getDatabase() + "/docs?id=" + UrlUtils_escapeDataString(data._id)

	request, err := NewHttpHead(url)
	if err != nil {
		return nil, err
	}

	if data._changeVector != nil {
		request.Header.Set("If-None-Match", *data._changeVector)
	}

	return request, nil
}

func HeadDocumentCommand_setResponse(cmd *RavenCommand, response String, fromCache bool) error {
	var res HiLoResult
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	cmd.result = &res
	return nil
}
