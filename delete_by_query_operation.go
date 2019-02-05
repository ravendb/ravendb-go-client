package ravendb

import (
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

func NewDeleteByQueryOperation(queryToDelete *IndexQuery, options *QueryOperationOptions) (*DeleteByQueryOperation, error) {
	if queryToDelete == nil {
		return nil, newIllegalArgumentError("QueryToDelete cannot be null")
	}
	return &DeleteByQueryOperation{
		_queryToDelete: queryToDelete,
		_options:       options,
	}, nil
}

func (o *DeleteByQueryOperation) GetCommand(store *DocumentStore, conventions *DocumentConventions, cache *HttpCache) (RavenCommand, error) {
	var err error
	o.Command, err = NewDeleteByIndexCommand(conventions, o._queryToDelete, o._options)
	return o.Command, err
}

var _ RavenCommand = &DeleteByIndexCommand{}

type DeleteByIndexCommand struct {
	RavenCommandBase

	_conventions   *DocumentConventions
	_queryToDelete *IndexQuery
	_options       *QueryOperationOptions

	Result *OperationIDResult
}

func NewDeleteByIndexCommand(conventions *DocumentConventions, queryToDelete *IndexQuery, options *QueryOperationOptions) (*DeleteByIndexCommand, error) {
	if options == nil {
		options = &QueryOperationOptions{}
	}
	cmd := &DeleteByIndexCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_conventions:   conventions,
		_queryToDelete: queryToDelete,
		_options:       options,
	}
	return cmd, nil
}

func (c *DeleteByIndexCommand) createRequest(node *ServerNode) (*http.Request, error) {
	_options := c._options

	url := node.URL + "/databases/" + node.Database + fmt.Sprintf("/queries?allowStale=%v", _options.allowStale)

	if _options.maxOpsPerSecond != 0 {
		url += "&maxOpsPerSec=" + strconv.Itoa(_options.maxOpsPerSecond)
	}

	url += fmt.Sprintf("&details=%v", _options.retrieveDetails)

	if _options.staleTimeout != 0 {
		url += "&staleTimeout=" + durationToTimeSpan(_options.staleTimeout)
	}

	m := jsonExtensionsWriteIndexQuery(c._conventions, c._queryToDelete)
	d, err := jsonMarshal(m)
	// TODO: return error instead?
	panicIf(err != nil, "jsonMarshal failed with %s", err)

	request, err := NewHttpDelete(url, d)
	if err != nil {
		return nil, err
	}
	return request, nil
}

func (c *DeleteByIndexCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}

	return jsonUnmarshal(response, &c.Result)
}
