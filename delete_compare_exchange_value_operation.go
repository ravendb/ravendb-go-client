package ravendb

import (
	"net/http"
	"reflect"
	"strconv"
)

var (
	_ IOperation = &DeleteCompareExchangeValueOperation{}
)

type DeleteCompareExchangeValueOperation struct {
	Command *RemoveCompareExchangeValueCommand

	_clazz reflect.Type
	_key   string
	_index int
}

func NewDeleteCompareExchangeValueOperation(clazz reflect.Type, key string, index int) *DeleteCompareExchangeValueOperation {
	return &DeleteCompareExchangeValueOperation{
		_clazz: clazz,
		_key:   key,
		_index: index,
	}
}

func (o *DeleteCompareExchangeValueOperation) GetCommand(store *IDocumentStore, conventions *DocumentConventions, cache *HttpCache) RavenCommand {
	o.Command = NewRemoveCompareExchangeValueCommand(o._clazz, o._key, o._index, conventions)
	return o.Command
}

var _ RavenCommand = &RemoveCompareExchangeValueCommand{}

type RemoveCompareExchangeValueCommand struct {
	RavenCommandBase

	_clazz       reflect.Type
	_key         string
	_index       int
	_conventions *DocumentConventions

	Result *CompareExchangeResult
}

func NewRemoveCompareExchangeValueCommand(clazz reflect.Type, key string, index int, conventions *DocumentConventions) *RemoveCompareExchangeValueCommand {
	// TODO: validation
	cmd := &RemoveCompareExchangeValueCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_clazz:       clazz,
		_key:         key,
		_index:       index,
		_conventions: conventions,
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *RemoveCompareExchangeValueCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/cmpxchg?key=" + c._key + "&index=" + strconv.Itoa(c._index)

	return NewHttpDelete(url, nil)
}

func (c *RemoveCompareExchangeValueCommand) SetResponse(response []byte, fromCache bool) error {
	var err error
	c.Result, err = parseCompareExchangeResultFromString(c._clazz, response, c._conventions)
	return err
}
