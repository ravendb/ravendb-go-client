package ravendb

import (
	"net/http"
)

var (
	_ IServerOperation = &GetDatabaseRecordOperation{}
)

type GetDatabaseRecordOperation struct {
	database string

	Command *GetDatabaseRecordCommand
}

func NewGetDatabaseRecordOperation(database string) *GetDatabaseRecordOperation {
	return &GetDatabaseRecordOperation{
		database: database,
	}
}

func (o *GetDatabaseRecordOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	o.Command = NewGetDatabaseRecordCommand(conventions, o.database)
	return o.Command, nil
}

var _ RavenCommand = &GetDatabaseRecordCommand{}

type GetDatabaseRecordCommand struct {
	RavenCommandBase

	conventions *DocumentConventions
	database    string

	Result *DatabaseRecordWithEtag
}

func NewGetDatabaseRecordCommand(conventions *DocumentConventions, database string) *GetDatabaseRecordCommand {
	cmd := &GetDatabaseRecordCommand{
		RavenCommandBase: NewRavenCommandBase(),

		conventions: conventions,
		database:    database,
	}
	return cmd
}

func (c *GetDatabaseRecordCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/admin/databases?name=" + c.database
	return newHttpGet(url)
}

func (c *GetDatabaseRecordCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		c.Result = nil
		return nil
	}

	return jsonUnmarshal(response, &c.Result)
}
