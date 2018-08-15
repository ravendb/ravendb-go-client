package ravendb

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type CreateDatabaseOperation struct {
	databaseRecord    *DatabaseRecord
	replicationFactor int
}

func NewCreateDatabaseOperation(databaseRecord *DatabaseRecord) *CreateDatabaseOperation {
	return NewCreateDatabaseOperationWithReplicationFactor(databaseRecord, 1)
}

func NewCreateDatabaseOperationWithReplicationFactor(databaseRecord *DatabaseRecord, replicationFactor int) *CreateDatabaseOperation {
	return &CreateDatabaseOperation{
		databaseRecord:    databaseRecord,
		replicationFactor: replicationFactor,
	}
}

func (o *CreateDatabaseOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	return NewCreateDatabaseCommand(conventions, o.databaseRecord, o.replicationFactor)
}

var (
	_ RavenCommand = &CreateDatabaseCommand{}
)

type CreateDatabaseCommand struct {
	*RavenCommandBase

	conventions       *DocumentConventions
	databaseRecord    *DatabaseRecord
	replicationFactor int
	databaseName      string

	Result *DatabasePutResult
}

func NewCreateDatabaseCommand(conventions *DocumentConventions, databaseRecord *DatabaseRecord, replicationFactor int) *CreateDatabaseCommand {
	panicIf(databaseRecord.DatabaseName == "", "databaseRecord.DatabaseName cannot be empty")
	cmd := &CreateDatabaseCommand{
		RavenCommandBase: NewRavenCommandBase(),

		conventions:       conventions,
		databaseRecord:    databaseRecord,
		replicationFactor: replicationFactor,
		databaseName:      databaseRecord.DatabaseName,
	}
	return cmd
}

func (c *CreateDatabaseCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/admin/databases?name=" + c.databaseName
	url += "&replicationFactor=" + strconv.Itoa(c.replicationFactor)

	js, err := json.Marshal(c.databaseRecord)
	if err != nil {
		return nil, err
	}
	return NewHttpPut(url, js)
}

func (c *CreateDatabaseCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}
	var res DatabasePutResult
	err := json.Unmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
