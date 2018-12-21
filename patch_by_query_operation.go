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

func NewPatchByQueryOperation(queryToUpdate string) *PatchByQueryOperation {
	return &PatchByQueryOperation{
		_queryToUpdate: NewIndexQuery(queryToUpdate),
	}
}

func (o *PatchByQueryOperation) GetCommand(store *IDocumentStore, conventions *DocumentConventions, cache *HttpCache) RavenCommand {
	o.Command = NewPatchByQueryCommand(conventions, o._queryToUpdate, o._options)
	return o.Command
}

var _ RavenCommand = &PatchByQueryCommand{}

type PatchByQueryCommand struct {
	RavenCommandBase

	_conventions   *DocumentConventions
	_queryToUpdate *IndexQuery
	_options       *QueryOperationOptions

	Result *OperationIDResult
}

func NewPatchByQueryCommand(conventions *DocumentConventions, queryToUpdate *IndexQuery, options *QueryOperationOptions) *PatchByQueryCommand {
	if options == nil {
		options = &QueryOperationOptions{}
	}
	cmd := &PatchByQueryCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_conventions:   conventions,
		_queryToUpdate: queryToUpdate,
		_options:       options,
	}
	return cmd
}

func (c *PatchByQueryCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	_options := c._options

	url := node.GetUrl() + "/databases/" + node.GetDatabase() + fmt.Sprintf("/queries?allowStale=%v", _options.allowStale)

	if _options.maxOpsPerSecond != 0 {
		url += "&maxOpsPerSec=" + strconv.Itoa(_options.maxOpsPerSecond)
	}

	url += fmt.Sprintf("&details=%v", _options.retrieveDetails)

	if _options.staleTimeout != 0 {
		url += "&staleTimeout=" + durationToTimeSpan(_options.staleTimeout)
	}

	q := JsonExtensions_writeIndexQuery(c._conventions, c._queryToUpdate)
	m := map[string]interface{}{
		"Query": q,
	}

	d, err := jsonMarshal(m)
	panicIf(err != nil, "jsonMarshal failed with %s", err)

	request, err := NewHttpPatch(url, d)
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
