package ravendb

import (
	"net/http"
	"reflect"
)

var (
	_ IOperation = &DeleteCompareExchangeValueOperation{}
)

type DeleteCompareExchangeValueOperation struct {
	Command *RemoveCompareExchangeValueCommand

	_clazz reflect.Type
	_key   string
	_index int64
}

func NewDeleteCompareExchangeValueOperation(clazz reflect.Type, key string, index int64) (*DeleteCompareExchangeValueOperation, error) {
	if stringIsEmpty(key) {
		return nil, newIllegalArgumentError("The kye argument must have value")
	}

	return &DeleteCompareExchangeValueOperation{
		_clazz: clazz,
		_key:   key,
		_index: index,
	}, nil
}

func (o *DeleteCompareExchangeValueOperation) GetCommand(store *DocumentStore, conventions *DocumentConventions, cache *httpCache) (RavenCommand, error) {
	var err error
	o.Command, err = NewRemoveCompareExchangeValueCommand(o._clazz, o._key, o._index, conventions)
	return o.Command, err
}

var _ RavenCommand = &RemoveCompareExchangeValueCommand{}

type RemoveCompareExchangeValueCommand struct {
	RavenCommandBase

	_clazz       reflect.Type
	_key         string
	_index       int64
	_conventions *DocumentConventions

	Result *CompareExchangeResult
}

func NewRemoveCompareExchangeValueCommand(clazz reflect.Type, key string, index int64, conventions *DocumentConventions) (*RemoveCompareExchangeValueCommand, error) {
	if stringIsEmpty(key) {
		return nil, newIllegalArgumentError("The kye argument must have value")
	}
	cmd := &RemoveCompareExchangeValueCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_clazz:       clazz,
		_key:         key,
		_index:       index,
		_conventions: conventions,
	}
	cmd.IsReadRequest = true
	return cmd, nil
}

func (c *RemoveCompareExchangeValueCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/cmpxchg?key=" + c._key + "&index=" + i64toa(c._index)

	return newHttpDelete(url, nil)
}

func (c *RemoveCompareExchangeValueCommand) setResponse(response []byte, fromCache bool) error {
	var err error
	c.Result, err = parseCompareExchangeResultFromString(c._clazz, response, c._conventions)
	return err
}
