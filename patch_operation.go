package ravendb

import (
	"encoding/json"
	"net/http"
)

var (
	_ IOperation = &PatchOperation{}
)

// Note: in Java it's Payload nested in PatchOperation
type PatchOperationPayload struct {
	patch          *PatchRequest
	patchIfMissing *PatchRequest
}

func NewPatchOperationPayload(patch *PatchRequest, patchIfMissing *PatchRequest) *PatchOperationPayload {
	return &PatchOperationPayload{
		patch:          patch,
		patchIfMissing: patchIfMissing,
	}
}

func (p *PatchOperationPayload) getPatch() *PatchRequest {
	return p.patch
}

func (p *PatchOperationPayload) getPatchIfMissing() *PatchRequest {
	return p.patchIfMissing
}

// Note: in Java it's Result nested in PatchOperation
type PatchOperationResult struct {
	Status   PatchStatus `json:"Status"`
	Document interface{} `json:"Document"`
}

func (r *PatchOperationResult) GetStatus() PatchStatus {
	return r.Status
}

func (r *PatchOperationResult) SetStatus(status PatchStatus) {
	r.Status = status
}

func (r *PatchOperationResult) GetDocument() interface{} {
	return r.Document
}

func (r *PatchOperationResult) SetDocument(document interface{}) {
	r.Document = document
}

type PatchOperation struct {
	Command *PatchCommand

	_id                              string
	_changeVector                    *string
	_patch                           *PatchRequest
	_patchIfMissing                  *PatchRequest
	_skipPatchIfChangeVectorMismatch bool
}

func NewPatchOperation(id string, changeVector *string, patch *PatchRequest, patchIfMissing *PatchRequest, skipPatchIfChangeVectorMismatch bool) *PatchOperation {
	panicIf(patch == nil, "Patch cannot be nil")
	panicIf(stringIsWhitespace(patch.GetScript()), "Patch script cannot be empty")
	panicIf(patchIfMissing != nil && stringIsWhitespace(patchIfMissing.GetScript()), "PatchIfMissing script cannot be empty")
	return &PatchOperation{
		_id:                              id,
		_changeVector:                    changeVector,
		_patch:                           patch,
		_patchIfMissing:                  patchIfMissing,
		_skipPatchIfChangeVectorMismatch: skipPatchIfChangeVectorMismatch,
	}
}

func (o *PatchOperation) GetCommand(store *IDocumentStore, conventions *DocumentConventions, cache *HttpCache) RavenCommand {
	o.Command = NewPatchCommand(conventions, o._id, o._changeVector, o._patch, o._patchIfMissing, o._skipPatchIfChangeVectorMismatch, false, false)
	return o.Command
}

var _ RavenCommand = &PatchCommand{}

type PatchCommand struct {
	RavenCommandBase

	_conventions                     *DocumentConventions
	_id                              string
	_changeVector                    *string
	_patch                           *PatchOperationPayload
	_skipPatchIfChangeVectorMismatch bool
	_returnDebugInformation          bool
	_test                            bool

	Result *PatchOperationResult
}

func NewPatchCommand(conventions *DocumentConventions, id string, changeVector *string,
	patch *PatchRequest, patchIfMissing *PatchRequest, skipPatchIfChangeVectorMismatch bool,
	returnDebugInformation bool, test bool) *PatchCommand {

	// TODO: validations

	cmd := &PatchCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_id:           id,
		_changeVector: changeVector,
		_patch:        NewPatchOperationPayload(patch, patchIfMissing),
		_skipPatchIfChangeVectorMismatch: skipPatchIfChangeVectorMismatch,
		_returnDebugInformation:          returnDebugInformation,
		_test: test,
	}

	return cmd
}

func (c *PatchCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/docs?id=" + UrlUtils_escapeDataString(c._id)

	if c._skipPatchIfChangeVectorMismatch {
		url += "&skipPatchIfChangeVectorMismatch=true"
	}

	if c._returnDebugInformation {
		url += "&debug=true"
	}

	if c._test {
		url += "&test=true"
	}

	patch := ObjectNode{}
	if c._patch.getPatch() != nil {
		patch = c._patch.getPatch().Serialize()
	}

	var patchIfMissing ObjectNode
	if c._patch.getPatchIfMissing() != nil {
		patchIfMissing = c._patch.getPatchIfMissing().Serialize()
	}

	m := map[string]interface{}{
		"Patch":          patch,
		"PatchIfMissing": patchIfMissing,
	}
	d, err := json.Marshal(m)
	panicIf(err != nil, "json.Marshal failed with %s", err)

	request, err := NewHttpPatch(url, d)
	if err != nil {
		return nil, err
	}
	addChangeVectorIfNotNull(c._changeVector, request)
	return request, nil
}

func (c *PatchCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return nil
	}

	return json.Unmarshal(response, &c.Result)
}
