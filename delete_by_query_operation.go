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

	queryToDelete *IndexQuery
	options       *QueryOperationOptions
}

func NewDeleteByQueryOperation(queryToDelete *IndexQuery, options *QueryOperationOptions) (*DeleteByQueryOperation, error) {
	if queryToDelete == nil {
		return nil, newIllegalArgumentError("QueryToDelete cannot be null")
	}
	return &DeleteByQueryOperation{
		queryToDelete: queryToDelete,
		options:       options,
	}, nil
}

func (o *DeleteByQueryOperation) GetCommand(store *DocumentStore, conventions *DocumentConventions, cache *httpCache) (RavenCommand, error) {
	var err error
	o.Command, err = NewDeleteByIndexCommand(conventions, o.queryToDelete, o.options)
	return o.Command, err
}

var _ RavenCommand = &DeleteByIndexCommand{}

type DeleteByIndexCommand struct {
	RavenCommandBase

	conventions   *DocumentConventions
	queryToDelete *IndexQuery
	options       *QueryOperationOptions

	Result *OperationIDResult
}

func NewDeleteByIndexCommand(conventions *DocumentConventions, queryToDelete *IndexQuery, options *QueryOperationOptions) (*DeleteByIndexCommand, error) {
	if options == nil {
		options = &QueryOperationOptions{}
	}
	cmd := &DeleteByIndexCommand{
		RavenCommandBase: NewRavenCommandBase(),

		conventions:   conventions,
		queryToDelete: queryToDelete,
		options:       options,
	}
	return cmd, nil
}

func (c *DeleteByIndexCommand) createRequest(node *ServerNode) (*http.Request, error) {
	options := c.options

	url := node.URL + "/databases/" + node.Database + fmt.Sprintf("/queries?allowStale=%v", options.allowStale)

	if options.maxOpsPerSecond != 0 {
		url += "&maxOpsPerSec=" + strconv.Itoa(options.maxOpsPerSecond)
	}

	url += fmt.Sprintf("&details=%v", options.retrieveDetails)

	if options.staleTimeout != 0 {
		url += "&staleTimeout=" + durationToTimeSpan(options.staleTimeout)
	}

	m := jsonExtensionsWriteIndexQuery(c.conventions, c.queryToDelete)
	d, err := jsonMarshal(m)
	// TODO: return error instead?
	panicIf(err != nil, "jsonMarshal failed with %s", err)

	request, err := newHttpDelete(url, d)
	if err != nil {
		return nil, err
	}
	return request, nil
}

func (c *DeleteByIndexCommand) setResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}

	return jsonUnmarshal(response, &c.Result)
}
