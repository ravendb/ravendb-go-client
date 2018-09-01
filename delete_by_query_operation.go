package ravendb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

var (
	_ IOperation = &DeleteByQueryOperation{}
)

type DeleteByQueryOperation struct {
	Command *DeleteByIndexCommand

	_queryToDelete *IndexQuery
	_options       *QueryOperationOptions
}

func NewDeleteByQueryOperation(queryToDelete *IndexQuery) *DeleteByQueryOperation {
	return NewDeleteByQueryOperationWithOptions(queryToDelete, nil)
}

func NewDeleteByQueryOperationWithOptions(queryToDelete *IndexQuery, options *QueryOperationOptions) *DeleteByQueryOperation {

	// TODO: validate queryToDelete
	return &DeleteByQueryOperation{
		_queryToDelete: queryToDelete,
		_options:       options,
	}
}

func (o *DeleteByQueryOperation) GetCommand(store *IDocumentStore, conventions *DocumentConventions, cache *HttpCache) RavenCommand {
	o.Command = NewDeleteByIndexCommand(conventions, o._queryToDelete, o._options)
	return o.Command
}

var _ RavenCommand = &DeleteByIndexCommand{}

type DeleteByIndexCommand struct {
	RavenCommandBase

	_conventions   *DocumentConventions
	_queryToDelete *IndexQuery
	_options       *QueryOperationOptions

	Result *OperationIdResult
}

func NewDeleteByIndexCommand(conventions *DocumentConventions, queryToDelete *IndexQuery, options *QueryOperationOptions) *DeleteByIndexCommand {
	if options == nil {
		options = NewQueryOperationOptions()
	}
	cmd := &DeleteByIndexCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_conventions:   conventions,
		_queryToDelete: queryToDelete,
		_options:       options,
	}
	return cmd
}

func (c *DeleteByIndexCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	_options := c._options

	url := node.GetUrl() + "/databases/" + node.GetDatabase() + fmt.Sprintf("/queries?allowStale=%v", _options.isAllowStale())

	if _options.getMaxOpsPerSecond() != 0 {
		url += "&maxOpsPerSec=" + strconv.Itoa(_options.getMaxOpsPerSecond())
	}

	url += fmt.Sprintf("&details=%v", _options.isRetrieveDetails())

	if _options.getStaleTimeout() != 0 {
		url += "&staleTimeout=" + durationToTimeSpan(_options.getStaleTimeout())
	}

	m := JsonExtensions_writeIndexQuery(c._conventions, c._queryToDelete)
	d, err := json.Marshal(m)
	// TODO: return error instead?
	panicIf(err != nil, "json.Marshal failed with %s", err)

	request, err := NewHttpDelete(url, d)
	if err != nil {
		return nil, err
	}
	return request, nil
}

func (c *DeleteByIndexCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		throwInvalidResponse()
	}

	return json.Unmarshal(response, &c.Result)
}
