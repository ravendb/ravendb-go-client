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

func NewGetCompareExchangeValuesOperationWithKeys(clazz reflect.Type, keys []string) *GetCompareExchangeValuesOperation {
	// TODO: validate
	return &GetCompareExchangeValuesOperation{
		_keys:  keys,
		_clazz: clazz,

		_start:     -1,
		_pageSize:  0,
		_startWith: "",
	}
}

func NewGetCompareExchangeValuesOperation(clazz reflect.Type, startWith string, start int, pageSize int) *GetCompareExchangeValuesOperation {
	return &GetCompareExchangeValuesOperation{
		_clazz: clazz,

		_start:     start,
		_pageSize:  pageSize,
		_startWith: startWith,
	}

}

var _ RavenCommand = &GetCompareExchangeValuesCommand{}

func (o *GetCompareExchangeValuesOperation) GetCommand(store *IDocumentStore, conventions *DocumentConventions, cache *HttpCache) RavenCommand {
	o.Command = NewGetCompareExchangeValuesCommand(o, conventions)
	return o.Command
}

type GetCompareExchangeValuesCommand struct {
	RavenCommandBase

	_operation   *GetCompareExchangeValuesOperation
	_conventions *DocumentConventions
	Result       map[string]*CompareExchangeValue
}

func NewGetCompareExchangeValuesCommand(operation *GetCompareExchangeValuesOperation, conventions *DocumentConventions) *GetCompareExchangeValuesCommand {
	cmd := &GetCompareExchangeValuesCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_operation:   operation,
		_conventions: conventions,
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *GetCompareExchangeValuesCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/cmpxchg?"

	if c._operation._keys != nil {
		for _, key := range c._operation._keys {
			url += "&key=" + UrlUtils_escapeDataString(key)
		}
	} else {
		if !stringIsEmpty(c._operation._startWith) {
			url += "&startsWith=" + UrlUtils_escapeDataString(c._operation._startWith)
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

func (c *GetCompareExchangeValuesCommand) SetResponse(response []byte, fromCache bool) error {
	res, err := CompareExchangeValueResultParser_getValues(c._operation._clazz, response, c._conventions)
	if err != nil {
		return err
	}
	c.Result = res
	return nil
}
