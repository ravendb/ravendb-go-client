package ravendb

import (
	"encoding/json"
	"net/http"
)

var (
	_ RavenCommand = &PutDocumentCommand{}
)

type PutDocumentCommand struct {
	*RavenCommandBase

	_id           string
	_changeVector *string
	_document     ObjectNode

	Result *PutResult
}

func NewPutDocumentCommand(id string, changeVector *string, document ObjectNode) *PutDocumentCommand {
	panicIf(id == "", "Id cannot be null")
	panicIf(document == nil, "document cannot be nil")

	cmd := &PutDocumentCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_id:           id,
		_changeVector: changeVector,
		_document:     document,
	}
	return cmd
}

func (c *PutDocumentCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/docs?id=" + urlEncode(c._id)

	d, err := json.Marshal(c._document)
	request, err := NewHttpPut(url, d)
	if err != nil {
		return nil, err
	}
	addChangeVectorIfNotNull(c._changeVector, request)
	return request, nil
}

func (c *PutDocumentCommand) SetResponse(response []byte, fromCache bool) error {
	var res PutResult
	err := json.Unmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
