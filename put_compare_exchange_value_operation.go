package ravendb

import (
	"net/http"
	"reflect"
)

var (
	_ IOperation = &PutCompareExchangeValueOperation{}
)

type PutCompareExchangeValueOperation struct {
	Command *PutCompareExchangeValueCommand

	_key   string
	_value interface{}
	_index int64
}

func NewPutCompareExchangeValueOperation(key string, value interface{}, index int64) (*PutCompareExchangeValueOperation, error) {
	if stringIsEmpty(key) {
		return nil, newIllegalArgumentError("The key argument must have value")
	}

	if index < 0 {
		return nil, newIllegalStateError("Index must be a non-negative number")
	}

	return &PutCompareExchangeValueOperation{
		_key:   key,
		_value: value,
		_index: index,
	}, nil
}

func (o *PutCompareExchangeValueOperation) GetCommand(store *DocumentStore, conventions *DocumentConventions, cache *HttpCache) (RavenCommand, error) {
	var err error
	o.Command, err = NewPutCompareExchangeValueCommand(o._key, o._value, o._index, conventions)
	return o.Command, err
}

var _ RavenCommand = &PutCompareExchangeValueCommand{}

type PutCompareExchangeValueCommand struct {
	RavenCommandBase

	_key         string
	_value       interface{}
	_index       int64
	_conventions *DocumentConventions

	Result *CompareExchangeResult
}

func NewPutCompareExchangeValueCommand(key string, value interface{}, index int64, conventions *DocumentConventions) (*PutCompareExchangeValueCommand, error) {
	if stringIsEmpty(key) {
		return nil, newIllegalArgumentError("The key argument must have value")
	}

	if index < 0 {
		return nil, newIllegalStateError("Index must be a non-negative number")
	}
	cmd := &PutCompareExchangeValueCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_key:         key,
		_value:       value,
		_index:       index,
		_conventions: conventions,
	}
	return cmd, nil
}

func (c *PutCompareExchangeValueCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/cmpxchg?key=" + c._key + "&index=" + i64toa(c._index)

	m := map[string]interface{}{
		"Object": c._value,
	}
	d, err := jsonMarshal(m)
	if err != nil {
		return nil, err
	}
	return NewHttpPut(url, d)

}

func (c *PutCompareExchangeValueCommand) setResponse(response []byte, fromCache bool) error {
	tp := reflect.TypeOf(c._value)
	res, err := parseCompareExchangeResultFromString(tp, response, c._conventions)
	if err != nil {
		return err
	}
	c.Result = res
	return nil
}
