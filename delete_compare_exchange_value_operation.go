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

func (o *DeleteCompareExchangeValueOperation) getCommand(store *IDocumentStore, conventions *DocumentConventions, cache *HttpCache) RavenCommand {
	o.Command = NewRemoveCompareExchangeValueCommand(o._clazz, o._key, o._index, conventions)
	return o.Command
}

var _ RavenCommand = &RemoveCompareExchangeValueCommand{}

type RemoveCompareExchangeValueCommand struct {
	*RavenCommandBase

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

func (c *RemoveCompareExchangeValueCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/cmpxchg?key=" + c._key + "&index=" + strconv.Itoa(c._index)

	return NewHttpDelete(url, nil)
}

func (c *RemoveCompareExchangeValueCommand) setResponse(response []byte, fromCache bool) error {
	res, err := CompareExchangeResult_parseFromString(c._clazz, response, c._conventions)
	if err != nil {
		return err
	}
	c.Result = res
	return nil
}