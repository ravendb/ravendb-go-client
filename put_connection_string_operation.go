package ravendb

import (
	"encoding/json"
	"net/http"
)

var (
	_ IMaintenanceOperation = &PutConnectionStringOperation{}
)

type PutConnectionStringOperation struct {
	_connectionString interface{}

	Command *PutConnectionStringCommand
}

func NewPutConnectionStringOperation(connectionString interface{}) *PutConnectionStringOperation {
	return &PutConnectionStringOperation{
		_connectionString: connectionString,
	}
}

func (o *PutConnectionStringOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewPutConnectionStringCommand(o._connectionString)
	return o.Command
}

var _ RavenCommand = &PutConnectionStringCommand{}

type PutConnectionStringCommand struct {
	*RavenCommandBase

	_connectionString interface{}

	Result *PutConnectionStringResult
}

func NewPutConnectionStringCommand(connectionString interface{}) *PutConnectionStringCommand {
	return &PutConnectionStringCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_connectionString: connectionString,
	}
}

func (c *PutConnectionStringCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/admin/connection-strings"

	d, err := json.Marshal(c._connectionString)
	if err != nil {
		// TODO: change err into RuntimeException() ?
		return nil, err
	}
	return NewHttpPut(url, d)
}

func (c *PutConnectionStringCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}

	var res PutConnectionStringResult
	err := json.Unmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
