package ravendb

import (
	"bytes"
	"encoding/json"
	"net/http"
)

var (
	_ RavenCommand = &PutDocumentCommand{}
)

type PutDocumentCommand struct {
	*RavenCommandBase

	_id           String
	_changeVector String
	_document     ObjectNode

	Result *PutResult
}

func NewPutDocumentCommand(id String, changeVector String, document ObjectNode) *PutDocumentCommand {
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
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/docs?id=" + urlEncode(c._id)

	d, err := json.Marshal(c._document)
	must(err)
	body := bytes.NewBuffer(d)
	// TODO: use NewPutRequest?
	request, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}
	// TODO: set Content-Type to application/json?
	addChangeVectorIfNotNull(c._changeVector, request)
	return request, nil
}

func (c *PutDocumentCommand) setResponse(response String, fromCache bool) error {
	var res PutResult
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
