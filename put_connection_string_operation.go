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

func (o *PutConnectionStringOperation) getCommand(conventions *DocumentConventions) RavenCommand {
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

func (c *PutConnectionStringCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/admin/connection-strings"

	d, err := json.Marshal(c._connectionString)
	if err != nil {
		// TODO: change err into RuntimeException() ?
		return nil, err
	}
	return NewHttpPut(url, string(d))
}

func (c *PutConnectionStringCommand) setResponse(response string, fromCache bool) error {
	if response == "" {
		return throwInvalidResponse()
	}

	var res PutConnectionStringResult
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
