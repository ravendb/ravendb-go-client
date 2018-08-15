package ravendb

import (
	"encoding/json"
	"net/http"
	"strconv"
)

var _ IMaintenanceOperation = &GetTermsOperation{}

type GetTermsOperation struct {
	_indexName string
	_field     string
	_fromValue string
	_pageSize  int // 0 for unset

	Command *GetTermsCommand
}

func NewGetTermsOperation(indexName string, field string, fromValue string) *GetTermsOperation {
	return NewGetTermsOperationWithPageSize(indexName, field, fromValue, 0)
}

func NewGetTermsOperationWithPageSize(indexName string, field string, fromValue string, pageSize int) *GetTermsOperation {
	panicIf(indexName == "", "Index name connot be empty")
	panicIf(field == "", "Field name connot be empty")
	return &GetTermsOperation{
		_indexName: indexName,
		_field:     field,
		_fromValue: fromValue,
		_pageSize:  pageSize,
	}
}

func (o *GetTermsOperation) getCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewGetTermsCommand(o._indexName, o._field, o._fromValue, o._pageSize)
	return o.Command
}

var (
	_ RavenCommand = &GetTermsCommand{}
)

type GetTermsCommand struct {
	*RavenCommandBase

	_indexName string
	_field     string
	_fromValue string
	_pageSize  int

	Result []string
}

func NewGetTermsCommand(indexName string, field string, fromValue string, pageSize int) *GetTermsCommand {
	panicIf(indexName == "", "Index name connot be empty")

	res := &GetTermsCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_indexName: indexName,
		_field:     field,
		_fromValue: fromValue,
		_pageSize:  pageSize,
	}
	res.IsReadRequest = true
	return res
}

func (c *GetTermsCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	pageSize := ""
	if c._pageSize > 0 {
		pageSize = strconv.Itoa(c._pageSize)
	}
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/indexes/terms?name=" + UrlUtils_escapeDataString(c._indexName) + "&field=" + UrlUtils_escapeDataString(c._field) + "&fromValue=" + c._fromValue + "&pageSize=" + pageSize

	return NewHttpGet(url)
}

func (c *GetTermsCommand) SetResponse(response []byte, fromCache bool) error {
	if response == nil {
		return throwInvalidResponse()
	}

	var res TermsQueryResult
	err := json.Unmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = res.getTerms()
	return nil
}
