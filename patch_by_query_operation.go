package ravendb

import (
	"encoding/json"
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

func (o *PatchByQueryOperation) getCommand(store *IDocumentStore, conventions *DocumentConventions, cache *HttpCache) RavenCommand {
	o.Command = NewPatchByQueryCommand(conventions, o._queryToUpdate, o._options)
	return o.Command
}

var _ RavenCommand = &PatchByQueryCommand{}

type PatchByQueryCommand struct {
	*RavenCommandBase

	_conventions   *DocumentConventions
	_queryToUpdate *IndexQuery
	_options       *QueryOperationOptions

	Result *OperationIdResult
}

func NewPatchByQueryCommand(conventions *DocumentConventions, queryToUpdate *IndexQuery, options *QueryOperationOptions) *PatchByQueryCommand {
	if options == nil {
		options = NewQueryOperationOptions()
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

	url := node.getUrl() + "/databases/" + node.getDatabase() + fmt.Sprintf("/queries?allowStale=%v", _options.isAllowStale())

	if _options.getMaxOpsPerSecond() != 0 {
		url += "&maxOpsPerSec=" + strconv.Itoa(_options.getMaxOpsPerSecond())
	}

	url += fmt.Sprintf("&details=%v", _options.isRetrieveDetails())

	if _options.getStaleTimeout() != 0 {
		url += "&staleTimeout=" + TimeUtils_durationToTimeSpan(_options.getStaleTimeout())
	}

	q := JsonExtensions_writeIndexQuery(c._conventions, c._queryToUpdate)
	m := map[string]Object{
		"Query": q,
	}

	d, err := json.Marshal(m)
	panicIf(err != nil, "json.Marshal failed with %s", err)

	request, err := NewHttpPatch(url, d)
	if err != nil {
		return nil, err
	}
	return request, nil
}

func (c *PatchByQueryCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		throwInvalidResponse()
	}

	var res OperationIdResult
	err := json.Unmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
