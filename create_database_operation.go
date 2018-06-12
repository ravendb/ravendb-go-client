package ravendb

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func NewCreateDatabaseOperation(databaseRecord *DatabaseRecord) *CreateDatabaseCommand {
	return NewCreateDatabaseOperationWithReplicationFactor(databaseRecord, 1)
}

func NewCreateDatabaseOperationWithReplicationFactor(databaseRecord *DatabaseRecord, replicationFactor int) *CreateDatabaseCommand {
	// TODO: convention is passed at getCommand() time
	return NewCreateDatabaseCommand(nil, databaseRecord, replicationFactor)
}

var (
	_ RavenCommand = &CreateDatabaseCommand{}
)

type CreateDatabaseCommand struct {
	*RavenCommandBase

	conventions       *DocumentConventions
	databaseRecord    *DatabaseRecord
	replicationFactor int
	databaseName      String

	Result *DatabasePutResult
}

func NewCreateDatabaseCommand(conventions *DocumentConventions, databaseRecord *DatabaseRecord, replicationFactor int) *CreateDatabaseCommand {
	panicIf(databaseRecord.DatabaseName == "", "databaseRecord.DatabaseName cannot be empty")
	cmd := &CreateDatabaseCommand{
		RavenCommandBase:  NewRavenCommandBase(),
		conventions:       conventions,
		databaseRecord:    databaseRecord,
		replicationFactor: replicationFactor,
		databaseName:      databaseRecord.DatabaseName,
	}
	return cmd
}

func (c *CreateDatabaseCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/admin/databases?name=" + c.databaseName
	url += "&replicationFactor=" + strconv.Itoa(c.replicationFactor)

	js, err := json.Marshal(c.databaseRecord)
	if err != nil {
		return nil, err
	}
	return NewHttpPut(url, string(js))
}

func (c *CreateDatabaseCommand) setResponse(response String, fromCache bool) error {
	if response == "" {
		return throwInvalidResponse()
	}
	var res DatabasePutResult
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
