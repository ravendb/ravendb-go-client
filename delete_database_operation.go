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

	Command *DeleteDatabaseCommand
}

type DeleteDatabaseParameters struct {
	DatabaseNames             []string       `json:"DatabaseNames"`
	HardDelete                bool           `json:"HardDelete"`
	FromNodes                 []string       `json:"FromNodes"`
	TimeToWaitForConfirmation *time.Duration `json:"TimeToWaitForConfirmation"`
}

func NewDeleteDatabasesOperation(databaseName string, hardDelete bool) *DeleteDatabasesOperation {
	return NewDeleteDatabasesOperation2(databaseName, hardDelete, "", 0)
}

func NewDeleteDatabasesOperation2(databaseName string, hardDelete bool, fromNode string, timeToWaitForConfirmation time.Duration) *DeleteDatabasesOperation {
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

func (o *DeleteDatabasesOperation) getCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewDeleteDatabaseCommand(conventions, o.parameters)
	return o.Command
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
		RavenCommandBase: NewRavenCommandBase(),

		parameters: string(d),
	}
	return cmd
}

func (c *DeleteDatabaseCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/admin/databases"
	return NewHttpDelete(url, c.parameters)
}

func (c *DeleteDatabaseCommand) setResponse(response []byte, fromCache bool) error {
	var res DeleteDatabaseResult
	err := json.Unmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
