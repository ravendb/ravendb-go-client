package ravendb

import (
	"fmt"
	"net/http"
	"reflect"
)

var (
	_ IOperation = &PatchOperation{}
)

// PatchOperationPayload represents payload of patch operation
// Note: in Java it's Payload nested in PatchOperation
type PatchOperationPayload struct {
	patch          *PatchRequest
	patchIfMissing *PatchRequest
}

// PatchOperationResult represents result of patch operation
// Note: in Java it's Result nested in PatchOperation
type PatchOperationResult struct {
	Status   PatchStatus            `json:"Status"`
	Document map[string]interface{} `json:"Document"`
}

func (r *PatchOperationResult) GetResult(result interface{}) error {
	fmt.Printf("Document: %v\n", r.Document)
	entityType := reflect.TypeOf(result)
	entity, err := makeStructFromJSONMap(entityType, r.Document)
	if err != nil {
		return err
	}
	setInterfaceToValue(result, entity)
	return nil
}

// PatchOperation represents patch operation
type PatchOperation struct {
	Command *PatchCommand

	id                              string
	changeVector                    *string
	patch                           *PatchRequest
	patchIfMissing                  *PatchRequest
	skipPatchIfChangeVectorMismatch bool
}

// NewPatchOperation returns new PatchOperation
func NewPatchOperation(id string, changeVector *string, patch *PatchRequest, patchIfMissing *PatchRequest, skipPatchIfChangeVectorMismatch bool) *PatchOperation {
	panicIf(patch == nil, "Patch cannot be nil")
	panicIf(stringIsWhitespace(patch.Script), "Patch script cannot be empty")
	panicIf(patchIfMissing != nil && stringIsWhitespace(patchIfMissing.Script), "PatchIfMissing script cannot be empty")
	return &PatchOperation{
		id:                              id,
		changeVector:                    changeVector,
		patch:                           patch,
		patchIfMissing:                  patchIfMissing,
		skipPatchIfChangeVectorMismatch: skipPatchIfChangeVectorMismatch,
	}
}

func (o *PatchOperation) GetCommand(store *DocumentStore, conventions *DocumentConventions, cache *HttpCache) RavenCommand {
	o.Command = NewPatchCommand(conventions, o.id, o.changeVector, o.patch, o.patchIfMissing, o.skipPatchIfChangeVectorMismatch, false, false)
	return o.Command
}

var _ RavenCommand = &PatchCommand{}

// PatchCommand represents patch command
type PatchCommand struct {
	RavenCommandBase

	// TODO: unused
	//conventions                     *DocumentConventions

	id                              string
	changeVector                    *string
	patch                           *PatchOperationPayload
	skipPatchIfChangeVectorMismatch bool
	returnDebugInformation          bool
	test                            bool

	Result *PatchResult
}

// NewPatchCommand returns new PatchCommand
func NewPatchCommand(conventions *DocumentConventions, id string, changeVector *string,
	patch *PatchRequest, patchIfMissing *PatchRequest, skipPatchIfChangeVectorMismatch bool,
	returnDebugInformation bool, test bool) *PatchCommand {

	// TODO: validations

	payload := &PatchOperationPayload{
		patch:          patch,
		patchIfMissing: patchIfMissing,
	}
	cmd := &PatchCommand{
		RavenCommandBase: NewRavenCommandBase(),

		id:                              id,
		changeVector:                    changeVector,
		patch:                           payload,
		skipPatchIfChangeVectorMismatch: skipPatchIfChangeVectorMismatch,
		returnDebugInformation:          returnDebugInformation,
		test:                            test,
	}

	return cmd
}

// CreateRequest creates http request
func (c *PatchCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/docs?id=" + urlUtilsEscapeDataString(c.id)

	if c.skipPatchIfChangeVectorMismatch {
		url += "&skipPatchIfChangeVectorMismatch=true"
	}

	if c.returnDebugInformation {
		url += "&debug=true"
	}

	if c.test {
		url += "&test=true"
	}

	patch := map[string]interface{}{}
	if c.patch.patch != nil {
		patch = c.patch.patch.Serialize()
	}

	var patchIfMissing map[string]interface{}
	if c.patch.patchIfMissing != nil {
		patchIfMissing = c.patch.patchIfMissing.Serialize()
	}

	m := map[string]interface{}{
		"Patch":          patch,
		"PatchIfMissing": patchIfMissing,
	}
	d, err := jsonMarshal(m)
	panicIf(err != nil, "jsonMarshal failed with %s", err)

	request, err := NewHttpPatch(url, d)
	if err != nil {
		return nil, err
	}
	addChangeVectorIfNotNull(c.changeVector, request)
	return request, nil
}

// SetResponse sets response
func (c *PatchCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return nil
	}

	return jsonUnmarshal(response, &c.Result)
}
