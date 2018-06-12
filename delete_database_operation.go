package ravendb

import (
	"encoding/json"
	"net/http"
	"time"
)

var (
	_ IMaintenanceOperation = &DeleteDatabasesOperation{}
)

type DeleteDatabasesOperation struct {
	parameters *DeleteDatabaseParameters
}

type DeleteDatabaseParameters struct {
	DatabaseNames             []string       `json:"DatabaseNames"`
	HardDelete                bool           `json:"HardDelete"`
	FromNodes                 []string       `json:"FromNodes"`
	TimeToWaitForConfirmation *time.Duration `json:"TimeToWaitForConfirmation"`
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
		parameters.FromNodes = []string{fromNode}
	}
	return NewDeleteDatabasesOperationWithParameters(parameters)
}

func NewDeleteDatabasesOperationWithParameters(parameters *DeleteDatabaseParameters) *DeleteDatabasesOperation {
	return &DeleteDatabasesOperation{
		parameters: parameters,
	}
}

func (o *DeleteDatabasesOperation) getRealCommand(conventions *DocumentConventions) *DeleteDatabaseCommand {
	return NewDeleteDatabaseCommand(conventions, o.parameters)
}

func (o *DeleteDatabasesOperation) getCommand(conventions *DocumentConventions) RavenCommand {
	return NewDeleteDatabaseCommand(conventions, o.parameters)
}

var _ RavenCommand = &DeleteDatabaseCommand{}

type DeleteDatabaseCommand struct {
	*RavenCommandBase

	parameters string

	Result *DeleteDatabaseResult
}

func NewDeleteDatabaseCommand(conventions *DocumentConventions, parameters *DeleteDatabaseParameters) *DeleteDatabaseCommand {
	d, err := json.Marshal(parameters)
	must(err)

	cmd := &DeleteDatabaseCommand{
		parameters: string(d),
	}
	return cmd
}

func (c *DeleteDatabaseCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/admin/databases"
	return NewHttpDelete(url, c.parameters)
}

func (c *DeleteDatabaseCommand) setResponse(response String, fromCache bool) error {
	var res DeleteDatabaseResult
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
