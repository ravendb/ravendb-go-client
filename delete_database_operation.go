package ravendb

import (
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

func (o *DeleteDatabasesOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewDeleteDatabaseCommand(conventions, o.parameters)
	return o.Command
}

var _ RavenCommand = &DeleteDatabaseCommand{}

type DeleteDatabaseCommand struct {
	RavenCommandBase

	parameters []byte

	Result *DeleteDatabaseResult
}

func NewDeleteDatabaseCommand(conventions *DocumentConventions, parameters *DeleteDatabaseParameters) *DeleteDatabaseCommand {
	d, err := jsonMarshal(parameters)
	must(err)

	cmd := &DeleteDatabaseCommand{
		RavenCommandBase: NewRavenCommandBase(),

		parameters: d,
	}
	return cmd
}

func (c *DeleteDatabaseCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/admin/databases"
	return NewHttpDelete(url, c.parameters)
}

func (c *DeleteDatabaseCommand) SetResponse(response []byte, fromCache bool) error {
	return jsonUnmarshal(response, &c.Result)
}
