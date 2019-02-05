package ravendb

import (
	"net/http"
	"strconv"
)

var _ IMaintenanceOperation = &GetIndexesOperation{}

type GetIndexesOperation struct {
	_start    int
	_pageSize int

	Command *GetIndexesCommand
}

func NewGetIndexesOperation(_start int, _pageSize int) *GetIndexesOperation {
	return &GetIndexesOperation{
		_start:    _start,
		_pageSize: _pageSize,
	}
}

func (o *GetIndexesOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	o.Command = NewGetIndexesCommand(o._start, o._pageSize)
	return o.Command, nil
}

var (
	_ RavenCommand = &GetIndexesCommand{}
)

type GetIndexesCommand struct {
	RavenCommandBase

	_start    int
	_pageSize int

	Result []*IndexDefinition
}

func NewGetIndexesCommand(_start int, _pageSize int) *GetIndexesCommand {
	res := &GetIndexesCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_start:    _start,
		_pageSize: _pageSize,
	}
	res.IsReadRequest = true
	return res
}

func (c *GetIndexesCommand) createRequest(node *ServerNode) (*http.Request, error) {
	start := strconv.Itoa(c._start)
	pageSize := strconv.Itoa(c._pageSize)

	url := node.URL + "/databases/" + node.Database + "/indexes?start=" + start + "&pageSize=" + pageSize

	return NewHttpGet(url)
}

func (c *GetIndexesCommand) SetResponse(response []byte, fromCache bool) error {
	if response == nil {
		return throwInvalidResponse()
	}

	var res struct {
		Results []*IndexDefinition `json:"Results"`
	}

	err := jsonUnmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = res.Results
	return nil
}
