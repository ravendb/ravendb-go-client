package ravendb

import (
	"net/http"
)

var (
	_ RavenCommand = &PutDocumentCommand{}
)

type PutDocumentCommand struct {
	RavenCommandBase

	_id           string
	_changeVector *string
	_document     map[string]interface{}

	Result *PutResult
}

func NewPutDocumentCommand(id string, changeVector *string, document map[string]interface{}) *PutDocumentCommand {
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

func (c *PutDocumentCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/docs?id=" + urlEncode(c._id)

	d, err := jsonMarshal(c._document)
	if err != nil {
		return nil, err
	}
	request, err := NewHttpPut(url, d)
	if err != nil {
		return nil, err
	}
	addChangeVectorIfNotNull(c._changeVector, request)
	return request, nil
}

func (c *PutDocumentCommand) SetResponse(response []byte, fromCache bool) error {
	return jsonUnmarshal(response, &c.Result)
}
