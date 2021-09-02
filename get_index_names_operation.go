package ravendb

import (
	"net/http"
	"strconv"
)

var _ IMaintenanceOperation = &GetIndexNamesOperation{}

type GetIndexNamesOperation struct {
	_start    int
	_pageSize int // 0 for unset

	Command *GetIndexNamesCommand
}

func NewGetIndexNamesOperation(start int, pageSize int) *GetIndexNamesOperation {
	return &GetIndexNamesOperation{
		_start:    start,
		_pageSize: pageSize,
	}
}

func (o *GetIndexNamesOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	o.Command = NewGetIndexNamesCommand(o._start, o._pageSize)
	return o.Command, nil
}

var (
	_ RavenCommand = &GetIndexNamesCommand{}
)

type GetIndexNamesCommand struct {
	RavenCommandBase

	_start    int
	_pageSize int // 0 for unset

	Result []string
}

func NewGetIndexNamesCommand(start int, pageSize int) *GetIndexNamesCommand {

	res := &GetIndexNamesCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_start:    start,
		_pageSize: pageSize,
	}
	res.IsReadRequest = true
	return res
}

func (c *GetIndexNamesCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	start := strconv.Itoa(c._start)
	pageSize := strconv.Itoa(c._pageSize)
	url := node.URL + "/databases/" + node.Database + "/indexes?start=" + start + "&pageSize=" + pageSize + "&namesOnly=true"

	return newHttpGet(url)
}

func (c *GetIndexNamesCommand) SetResponse(response []byte, fromCache bool) error {
	if response == nil {
		return throwInvalidResponse()
	}

	var res struct {
		Results []string `json:"Results"`
	}
	err := jsonUnmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = res.Results
	return nil
}
