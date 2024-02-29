package ravendb

import (
	"fmt"
	"net/http"
	"strconv"
)

var (
	_ IOperation = &PatchByQueryOperation{}
)

type PatchByQueryOperation struct {
	Command *PatchByQueryCommand

	_queryToUpdate *IndexQuery
	_options       *QueryOperationOptions
}


func NewPatchByQueryOperation(queryToUpdate string, opts ...*QueryOperationOptions) *PatchByQueryOperation {
	op := &PatchByQueryOperation{
		_queryToUpdate: NewIndexQuery(queryToUpdate),
	}

	if len(opts) > 0 && opts[0] != nil {
		op._options = opts[0]
	}

	return p
}

func (o *PatchByQueryOperation) GetCommand(store *DocumentStore, conventions *DocumentConventions, cache *httpCache) (RavenCommand, error) {
	var err error
	o.Command, err = NewPatchByQueryCommand(conventions, o._queryToUpdate, o._options)
	return o.Command, err
}

var _ RavenCommand = &PatchByQueryCommand{}

type PatchByQueryCommand struct {
	RavenCommandBase

	_conventions   *DocumentConventions
	_queryToUpdate *IndexQuery
	_options       *QueryOperationOptions

	Result *OperationIDResult
}

func NewPatchByQueryCommand(conventions *DocumentConventions, queryToUpdate *IndexQuery, options *QueryOperationOptions) (*PatchByQueryCommand, error) {
	if queryToUpdate == nil {
		return nil, newIllegalArgumentError("QueryToUpdate cannot be null")
	}

	if options == nil {
		options = &QueryOperationOptions{}
	}
	cmd := &PatchByQueryCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_conventions:   conventions,
		_queryToUpdate: queryToUpdate,
		_options:       options,
	}
	return cmd, nil
}

func (c *PatchByQueryCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	_options := c._options

	url := node.URL + "/databases/" + node.Database + fmt.Sprintf("/queries?allowStale=%v", _options.AllowStale)

	if _options.maxOpsPerSecond != 0 {
		url += "&maxOpsPerSec=" + strconv.Itoa(_options.maxOpsPerSecond)
	}

	url += fmt.Sprintf("&details=%v", _options.retrieveDetails)

	if _options.staleTimeout != 0 {
		url += "&staleTimeout=" + durationToTimeSpan(_options.staleTimeout)
	}

	q := jsonExtensionsWriteIndexQuery(c._conventions, c._queryToUpdate)
	m := map[string]interface{}{
		"Query": q,
	}

	d, err := jsonMarshal(m)
	panicIf(err != nil, "jsonMarshal failed with %s", err)

	request, err := newHttpPatch(url, d)
	if err != nil {
		return nil, err
	}
	return request, nil
}

func (c *PatchByQueryCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}

	return jsonUnmarshal(response, &c.Result)
}
