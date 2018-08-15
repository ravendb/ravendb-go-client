package ravendb

import (
	"net/http"
)

var (
	_ IVoidMaintenanceOperation = &CreateSampleDataOperation{}
)

type CreateSampleDataOperation struct {
	Command *CreateSampleDataCommand
}

func NewCreateSampleDataOperation() *CreateSampleDataOperation {
	return &CreateSampleDataOperation{}
}

func (o *CreateSampleDataOperation) getCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewCreateSampleDataCommand(conventions)
	return o.Command
}

var _ RavenCommand = &CreateSampleDataCommand{}

type CreateSampleDataCommand struct {
	*RavenCommandBase
}

func NewCreateSampleDataCommand(conventions *DocumentConventions) *CreateSampleDataCommand {
	cmd := &CreateSampleDataCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}
	cmd.RavenCommandBase.responseType = RavenCommandResponseType_EMPTY
	return cmd
}

func (c *CreateSampleDataCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/studio/sample-data"

	return NewHttpPost(url, nil)
}
