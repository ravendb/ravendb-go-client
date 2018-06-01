package ravendb

import (
	"encoding/json"
	"net/http"
	"strconv"
)

func NewCreateDatabaseOperation(databaseRecord *DatabaseRecord) *RavenCommand {
	return NewCreateDatabaseOperationWithReplicationFactor(databaseRecord, 1)
}

func NewCreateDatabaseOperationWithReplicationFactor(databaseRecord *DatabaseRecord, replicationFactor int) *RavenCommand {
	// TODO: convention is passed at getCommand() time
	return NewCreateDatabaseCommand(nil, databaseRecord, replicationFactor)
}

type _CreateDatabaseCommand struct {
	conventions       *DocumentConventions
	databaseRecord    *DatabaseRecord
	replicationFactor int
	databaseName      String
}

func NewCreateDatabaseCommand(conventions *DocumentConventions, databaseRecord *DatabaseRecord, replicationFactor int) *RavenCommand {
	panicIf(databaseRecord.DatabaseName == "", "databaseRecord.DatabaseName cannot be empty")
	data := &_CreateDatabaseCommand{
		conventions:       conventions,
		databaseRecord:    databaseRecord,
		replicationFactor: replicationFactor,
		databaseName:      databaseRecord.DatabaseName,
	}

	cmd := NewRavenCommand()
	cmd.data = data
	cmd.createRequestFunc = CreateDatabaseCommand_createRequest
	cmd.setResponseFunc = CreateDatabaseCommand_setResponse
	return cmd
}

func CreateDatabaseCommand_createRequest(cmd *RavenCommand, node *ServerNode) (*http.Request, string) {
	data := cmd.data.(*_CreateDatabaseCommand)
	url := node.getUrl() + "/admin/databases?name=" + data.databaseName
	url += "&replicationFactor=" + strconv.Itoa(data.replicationFactor)

	js, err := json.Marshal(data.databaseRecord)
	must(err)
	request := NewHttpPut(url, string(js))
	return request, url
}

func CreateDatabaseCommand_setResponse(cmd *RavenCommand, response String, fromCache bool) error {
	if response == "" {
		return throwInvalidResponse()
	}
	var res DatabasePutResult
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	cmd.result = &res
	return nil
}
