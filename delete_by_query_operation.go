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
	Command *DeleteByQueryCommand

	_queryToDelete *IndexQuery
	_options       *QueryOperationOptions
}

func NewDeleteByQueryOperation(queryToDelete string) *DeleteByQueryOperation {
	return &DeleteByQueryOperation{
		_queryToDelete: NewIndexQuery(queryToDelete),
	}
}

func (o *DeleteByQueryOperation) getCommand(store *IDocumentStore, conventions *DocumentConventions, cache *HttpCache) RavenCommand {
	o.Command = NewDeleteByQueryCommand(conventions, o._queryToDelete, o._options)
	return o.Command
}

var _ RavenCommand = &DeleteByQueryCommand{}

type DeleteByQueryCommand struct {
	*RavenCommandBase

	_conventions   *DocumentConventions
	_queryToDelete *IndexQuery
	_options       *QueryOperationOptions

	Result *OperationIdResult
}

func NewDeleteByQueryCommand(conventions *DocumentConventions, queryToDelete *IndexQuery, options *QueryOperationOptions) *DeleteByQueryCommand {
	if options == nil {
		options = NewQueryOperationOptions()
	}
	cmd := &DeleteByQueryCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_conventions:   conventions,
		_queryToDelete: queryToDelete,
		_options:       options,
	}
	return cmd
}

func (c *DeleteByQueryCommand) createRequest(node *ServerNode) (*http.Request, error) {
	_options := c._options

	url := node.getUrl() + "/databases/" + node.getDatabase() + fmt.Sprintf("/queries?allowStale=%v", _options.isAllowStale())

	if _options.getMaxOpsPerSecond() != 0 {
		url += "&maxOpsPerSec=" + strconv.Itoa(_options.getMaxOpsPerSecond())
	}

	url += fmt.Sprintf("&details=%v", _options.isRetrieveDetails())

	if _options.getStaleTimeout() != 0 {
		url += "&staleTimeout=" + TimeUtils_durationToTimeSpan(_options.getStaleTimeout())
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

func (c *DeleteByQueryCommand) setResponse(response []byte, fromCache bool) error {
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
