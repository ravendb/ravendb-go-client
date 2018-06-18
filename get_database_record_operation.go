package ravendb

import (
	"encoding/json"
	"net/http"
)

var (
	_ IServerOperation = &GetDatabaseRecordOperation{}
)

type GetDatabaseRecordOperation struct {
	_database string

	Command *GetDatabaseRecordCommand
}

func NewGetDatabaseRecordOperation(database string) *GetDatabaseRecordOperation {
	return &GetDatabaseRecordOperation{
		_database: database,
	}
}

func (o *GetDatabaseRecordOperation) getCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewGetDatabaseRecordCommand(conventions, o._database)
	return o.Command
}

var _ RavenCommand = &GetDatabaseRecordCommand{}

type GetDatabaseRecordCommand struct {
	*RavenCommandBase

	_conventions *DocumentConventions
	_database    string

	Result *DatabaseRecordWithEtag
}

func NewGetDatabaseRecordCommand(conventions *DocumentConventions, database string) *GetDatabaseRecordCommand {
	cmd := &GetDatabaseRecordCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_conventions: conventions,
		_database:    database,
	}
	return cmd
}

func (c *GetDatabaseRecordCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/admin/databases?name=" + c._database
	return NewHttpGet(url)
}

func (c *GetDatabaseRecordCommand) setResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		c.Result = nil
		return nil
	}

	var res DatabaseRecordWithEtag
	err := json.Unmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
