package ravendb

import (
	"net/http"
	"reflect"
	"strconv"
)

var (
	_ IOperation = &PutCompareExchangeValueOperation{}
)

type PutCompareExchangeValueOperation struct {
	Command *PutCompareExchangeValueCommand

	_key   string
	_value interface{}
	_index int
}

func NewPutCompareExchangeValueOperation(key string, value interface{}, index int) *PutCompareExchangeValueOperation {
	return &PutCompareExchangeValueOperation{
		_key:   key,
		_value: value,
		_index: index,
	}
}

func (o *PutCompareExchangeValueOperation) GetCommand(store *IDocumentStore, conventions *DocumentConventions, cache *HttpCache) RavenCommand {
	o.Command = NewPutCompareExchangeValueCommand(o._key, o._value, o._index, conventions)
	return o.Command
}

var _ RavenCommand = &PutCompareExchangeValueCommand{}

type PutCompareExchangeValueCommand struct {
	RavenCommandBase

	_key         string
	_value       interface{}
	_index       int
	_conventions *DocumentConventions

	Result *CompareExchangeResult
}

func NewPutCompareExchangeValueCommand(key string, value interface{}, index int, conventions *DocumentConventions) *PutCompareExchangeValueCommand {
	// TODO: validation
	cmd := &PutCompareExchangeValueCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_key:         key,
		_value:       value,
		_index:       index,
		_conventions: conventions,
	}
	return cmd
}

func (c *PutCompareExchangeValueCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/cmpxchg?key=" + c._key + "&index=" + strconv.Itoa(c._index)

	m := map[string]interface{}{
		"Object": c._value,
	}
	d, err := jsonMarshal(m)
	if err != nil {
		return nil, err
	}
	return NewHttpPut(url, d)

}

func (c *PutCompareExchangeValueCommand) SetResponse(response []byte, fromCache bool) error {
	tp := reflect.TypeOf(c._value)
	res, err := CompareExchangeResult_parseFromString(tp, response, c._conventions)
	if err != nil {
		return err
	}
	c.Result = res
	return nil
}
