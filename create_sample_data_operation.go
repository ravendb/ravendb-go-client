package ravendb

import (
	"net/http"
)

// Note: this is only used for tests but a RavenCommand cannot be implemented outside
// of ravendb package

var (
	_ IVoidMaintenanceOperation = &CreateSampleDataOperation{}
)

// CreateSampleDataOperation represents operation to create sample data
type CreateSampleDataOperation struct {
	Command *CreateSampleDataCommand
}

// NewCreateSampleDataOperation
func NewCreateSampleDataOperation() *CreateSampleDataOperation {
	return &CreateSampleDataOperation{}
}

// GetCommand returns a comman
func (o *CreateSampleDataOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	o.Command = NewCreateSampleDataCommand(conventions)
	return o.Command, nil
}

var _ RavenCommand = &CreateSampleDataCommand{}

// CreateSampleDataCommand represents command for creating sample data
type CreateSampleDataCommand struct {
	RavenCommandBase
}

// NewCreateSampleDataCommand returns new CreateSampleDataCommand
func NewCreateSampleDataCommand(conventions *DocumentConventions) *CreateSampleDataCommand {
	cmd := &CreateSampleDataCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}
	cmd.RavenCommandBase.ResponseType = RavenCommandResponseTypeEmpty
	return cmd
}

func (c *CreateSampleDataCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/studio/sample-data"

	return newHttpPost(url, nil)
}
