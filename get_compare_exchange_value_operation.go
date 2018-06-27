package ravendb

import (
	"net/http"
	"reflect"
)

var (
	_ IOperation = &GetCompareExchangeValueOperation{}
)

type GetCompareExchangeValueOperation struct {
	Command *GetCompareExchangeValueCommand

	_key   string
	_clazz reflect.Type
}

func NewGetCompareExchangeValueOperation(clazz reflect.Type, key string) *GetCompareExchangeValueOperation {
	return &GetCompareExchangeValueOperation{
		_clazz: clazz,
		_key:   key,
	}
}

func (o *GetCompareExchangeValueOperation) getCommand(store *IDocumentStore, conventions *DocumentConventions, cache *HttpCache) RavenCommand {
	o.Command = NewGetCompareExchangeValueCommand(o._clazz, o._key, conventions)
	return o.Command
}

var _ RavenCommand = &GetCompareExchangeValueCommand{}

type GetCompareExchangeValueCommand struct {
	*RavenCommandBase

	_key         string
	_clazz       reflect.Type
	_conventions *DocumentConventions

	Result *CompareExchangeValue
}

func NewGetCompareExchangeValueCommand(clazz reflect.Type, key string, conventions *DocumentConventions) *GetCompareExchangeValueCommand {
	// TODO: validation
	cmd := &GetCompareExchangeValueCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_clazz:       clazz,
		_key:         key,
		_conventions: conventions,
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *GetCompareExchangeValueCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/cmpxchg?key=" + urlEncode(c._key)
	return NewHttpGet(url)

}

func (c *GetCompareExchangeValueCommand) setResponse(response []byte, fromCache bool) error {
	res, err := CompareExchangeValueResultParser_getValue(c._clazz, response, c._conventions)
	if err != nil {
		return err
	}
	c.Result = res
	return nil
}
