package ravendb

import (
	"encoding/json"
	"net/http"
	"time"
)

type DeleteDatabasesOperation struct {
	parameters *DeleteDatabaseParameters
}

type DeleteDatabaseParameters struct {
	DatabaseNames             []string       `json:"DatabaseNames"`
	HardDelete                bool           `json:"HardDelete"`
	FromNodes                 *[]string      `json:"FromNodes",omitempty`
	TimeToWaitForConfirmation *time.Duration `json:"TimeToWaitForConfirmation",omitempty`
}

func NewDeleteDatabasesOperation(databaseName String, hardDelete bool) *DeleteDatabasesOperation {
	return NewDeleteDatabasesOperation2(databaseName, hardDelete, "", 0)
}

func NewDeleteDatabasesOperation2(databaseName String, hardDelete bool, fromNode String, timeToWaitForConfirmation time.Duration) *DeleteDatabasesOperation {
	parameters := &DeleteDatabaseParameters{
		DatabaseNames: []string{databaseName},
		HardDelete:    hardDelete,
	}
	if timeToWaitForConfirmation != 0 {
		parameters.TimeToWaitForConfirmation = &timeToWaitForConfirmation
	}
	if fromNode != "" {
		parameters.FromNodes = &[]string{fromNode}
	}
	return NewDeleteDatabasesOperationWithParameters(parameters)
}

func NewDeleteDatabasesOperationWithParameters(parameters *DeleteDatabaseParameters) *DeleteDatabasesOperation {
	return &DeleteDatabasesOperation{
		parameters: parameters,
	}
}

func (o *DeleteDatabasesOperation) getCommand(conventions *DocumentConventions) *RavenCommand {
	return NewDeleteDatabaseCommand(conventions, o.parameters)
}

type _DeleteDatabaseCommand struct {
	parameters string
}

func NewDeleteDatabaseCommand(conventions *DocumentConventions, parameters *DeleteDatabaseParameters) *RavenCommand {
	d, err := json.Marshal(parameters)
	must(err)

	data := &_DeleteDatabaseCommand{
		parameters: string(d),
	}
	cmd := NewRavenCommand()
	cmd.data = data
	cmd.createRequestFunc = DeleteDatabaseCommand_createRequest
	cmd.setResponseFunc = DeleteDatabaseCommand_setResponse
	return cmd
}

func DeleteDatabaseCommand_createRequest(cmd *RavenCommand, node *ServerNode) (*http.Request, string) {
	data := cmd.data.(*_DeleteDatabaseCommand)
	url := node.getUrl() + "/admin/databases"
	return NewHttpDelete(url, data.parameters), url
}

func DeleteDatabaseCommand_setResponse(cmd *RavenCommand, response String, fromCache bool) error {
	var res DeleteDatabaseResult
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	cmd.result = &res
	return nil
}
