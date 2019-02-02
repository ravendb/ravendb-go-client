package ravendb

import (
	"net/http"
	"strconv"
)

var _ IMaintenanceOperation = &GetTermsOperation{}

// GetTermsOperation represents "get terms" operation
type GetTermsOperation struct {
	_indexName string
	_field     string
	_fromValue string
	_pageSize  int // 0 for unset

	Command *GetTermsCommand
}

// NewGetTermsOperation returns GetTermsOperation. pageSize 0 means default size
func NewGetTermsOperation(indexName string, field string, fromValue string, pageSize int) (*GetTermsOperation, error) {
	if indexName == "" {
		return nil, newIllegalStateError("Index name connot be empty")
	}
	if field == "" {
		return nil, newIllegalStateError("Field name connot be empty")

	}
	return &GetTermsOperation{
		_indexName: indexName,
		_field:     field,
		_fromValue: fromValue,
		_pageSize:  pageSize,
	}, nil
}

// GetCommand returns command for this operation
func (o *GetTermsOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	var err error
	o.Command, err = NewGetTermsCommand(o._indexName, o._field, o._fromValue, o._pageSize)
	if err != nil {
		return nil, err
	}
	return o.Command, err
}

var (
	_ RavenCommand = &GetTermsCommand{}
)

// GetTermsCommand represents "get terms" command
type GetTermsCommand struct {
	RavenCommandBase

	_indexName string
	_field     string
	_fromValue string
	_pageSize  int

	Result []string
}

// NewGetTermsCommand returns new GetTermsCommand
func NewGetTermsCommand(indexName string, field string, fromValue string, pageSize int) (*GetTermsCommand, error) {
	if indexName == "" {
		return nil, newIllegalArgumentError("Index name connot be empty")
	}

	res := &GetTermsCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_indexName: indexName,
		_field:     field,
		_fromValue: fromValue,
		_pageSize:  pageSize,
	}
	res.IsReadRequest = true
	return res, nil
}

// CreateRequest creates a request
func (c *GetTermsCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	pageSize := ""
	if c._pageSize > 0 {
		pageSize = strconv.Itoa(c._pageSize)
	}
	url := node.URL + "/databases/" + node.Database + "/indexes/terms?name=" + urlUtilsEscapeDataString(c._indexName) + "&field=" + urlUtilsEscapeDataString(c._field) + "&fromValue=" + c._fromValue + "&pageSize=" + pageSize

	return NewHttpGet(url)
}

// SetResponse sets a response
func (c *GetTermsCommand) SetResponse(response []byte, fromCache bool) error {
	if response == nil {
		return throwInvalidResponse()
	}

	var res TermsQueryResult
	err := jsonUnmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = res.Terms
	return nil
}
