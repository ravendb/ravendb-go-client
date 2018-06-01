package ravendb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type _PutDocumentCommand struct {
	_id           String
	_changeVector String
	_document     ObjectNode
}

func NewPutDocumentCommand(id String, changeVector String, document ObjectNode) *RavenCommand {
	panicIf(id == "", "Id cannot be null")
	panicIf(document == nil, "document cannot be nil")

	data := &_PutDocumentCommand{
		_id:           id,
		_changeVector: changeVector,
		_document:     document,
	}
	cmd := NewRavenCommand()
	cmd.data = data
	cmd.createRequestFunc = PutDocumentCommand_createRequest
	cmd.setResponseFunc = PutDocumentCommand_setResponse
	return cmd
}

func PutDocumentCommand_createRequest(cmd *RavenCommand, node *ServerNode) (*http.Request, string) {
	data := cmd.data.(*_PutDocumentCommand)

	url := node.getUrl() + "/databases/" + node.getDatabase() + "/docs?id=" + urlEncode(data._id)

	d, err := json.Marshal(data._document)
	must(err)
	body := bytes.NewBuffer(d)
	request, err := http.NewRequest(http.MethodPut, url, body)
	panicIf(err != nil, "http.NewRequest failed with %s", err)
	// TODO: set Content-Type to application/json?

	addChangeVectorIfNotNull(data._changeVector, request)
	return request, url
}

func PutDocumentCommand_setResponse(cmd *RavenCommand, response String, fromCache bool) error {
	var res PutResult
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	cmd.result = &res
	return nil
}

func addChangeVectorIfNotNull(changeVector string, req *http.Request) {
	if changeVector != "" {
		req.Header.Add("If-Match", fmt.Sprintf(`"%s"`, changeVector))
	}
}
