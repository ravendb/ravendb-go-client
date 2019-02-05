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

func NewGetCompareExchangeValueOperation(clazz reflect.Type, key string) (*GetCompareExchangeValueOperation, error) {
	if stringIsEmpty(key) {
		return nil, newIllegalArgumentError("The key argument must have value")
	}

	return &GetCompareExchangeValueOperation{
		_clazz: clazz,
		_key:   key,
	}, nil
}

func (o *GetCompareExchangeValueOperation) GetCommand(store *DocumentStore, conventions *DocumentConventions, cache *HttpCache) (RavenCommand, error) {
	var err error
	o.Command, err = NewGetCompareExchangeValueCommand(o._clazz, o._key, conventions)
	return o.Command, err
}

var _ RavenCommand = &GetCompareExchangeValueCommand{}

type GetCompareExchangeValueCommand struct {
	RavenCommandBase

	_key         string
	_clazz       reflect.Type
	_conventions *DocumentConventions

	Result *CompareExchangeValue
}

func NewGetCompareExchangeValueCommand(clazz reflect.Type, key string, conventions *DocumentConventions) (*GetCompareExchangeValueCommand, error) {
	if stringIsEmpty(key) {
		return nil, newIllegalArgumentError("The key argument must have value")
	}

	cmd := &GetCompareExchangeValueCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_clazz:       clazz,
		_key:         key,
		_conventions: conventions,
	}
	cmd.IsReadRequest = true
	return cmd, nil
}

func (c *GetCompareExchangeValueCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/cmpxchg?key=" + urlEncode(c._key)
	return NewHttpGet(url)

}

func (c *GetCompareExchangeValueCommand) setResponse(response []byte, fromCache bool) error {
	res, err := compareExchangeValueResultParserGetValue(c._clazz, response, c._conventions)
	if err != nil {
		return err
	}
	c.Result = res
	return nil
}
