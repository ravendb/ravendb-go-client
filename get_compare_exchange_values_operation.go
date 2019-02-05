package ravendb

import (
	"net/http"
	"reflect"
	"strconv"
)

var (
	_ IOperation = &GetCompareExchangeValuesOperation{}
)

type GetCompareExchangeValuesOperation struct {
	Command *GetCompareExchangeValuesCommand

	_clazz reflect.Type
	_keys  []string

	_startWith string
	_start     int // -1 for unset
	_pageSize  int
}

func NewGetCompareExchangeValuesOperationWithKeys(clazz reflect.Type, keys []string) (*GetCompareExchangeValuesOperation, error) {
	if len(keys) == 0 {
		return nil, newIllegalArgumentError("Keys cannot be null or empty array")
	}
	return &GetCompareExchangeValuesOperation{
		_keys:  keys,
		_clazz: clazz,

		_start:     -1,
		_pageSize:  0,
		_startWith: "",
	}, nil
}

func NewGetCompareExchangeValuesOperation(clazz reflect.Type, startWith string, start int, pageSize int) (*GetCompareExchangeValuesOperation, error) {
	return &GetCompareExchangeValuesOperation{
		_clazz: clazz,

		_start:     start,
		_pageSize:  pageSize,
		_startWith: startWith,
	}, nil
}

var _ RavenCommand = &GetCompareExchangeValuesCommand{}

func (o *GetCompareExchangeValuesOperation) GetCommand(store *DocumentStore, conventions *DocumentConventions, cache *HttpCache) (RavenCommand, error) {
	var err error
	o.Command, err = NewGetCompareExchangeValuesCommand(o, conventions)
	return o.Command, err
}

type GetCompareExchangeValuesCommand struct {
	RavenCommandBase

	_operation   *GetCompareExchangeValuesOperation
	_conventions *DocumentConventions
	Result       map[string]*CompareExchangeValue
}

func NewGetCompareExchangeValuesCommand(operation *GetCompareExchangeValuesOperation, conventions *DocumentConventions) (*GetCompareExchangeValuesCommand, error) {
	cmd := &GetCompareExchangeValuesCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_operation:   operation,
		_conventions: conventions,
	}
	cmd.IsReadRequest = true
	return cmd, nil
}

func (c *GetCompareExchangeValuesCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/cmpxchg?"

	if c._operation._keys != nil {
		for _, key := range c._operation._keys {
			url += "&key=" + urlUtilsEscapeDataString(key)
		}
	} else {
		if !stringIsEmpty(c._operation._startWith) {
			url += "&startsWith=" + urlUtilsEscapeDataString(c._operation._startWith)
		}

		if c._operation._start >= 0 {
			url += "&start=" + strconv.Itoa(c._operation._start)
		}

		if c._operation._pageSize > 0 {
			url += "&pageSize=" + strconv.Itoa(c._operation._pageSize)
		}
	}

	return NewHttpGet(url)
}

func (c *GetCompareExchangeValuesCommand) setResponse(response []byte, fromCache bool) error {
	res, err := compareExchangeValueResultParserGetValues(c._operation._clazz, response, c._conventions)
	if err != nil {
		return err
	}
	c.Result = res
	return nil
}
