package ravendb

import (
	"net/http"
)

var (
	_ IMaintenanceOperation = &ConfigureRevisionsOperation{}
)

// ConfigureRevisionsOperation represents configure revisions operation
type ConfigureRevisionsOperation struct {
	_configuration *RevisionsConfiguration
	Command        *ConfigureRevisionsCommand
}

// NewConfigureRevisionsOperation returns new ConfigureRevisionsOperation
func NewConfigureRevisionsOperation(configuration *RevisionsConfiguration) *ConfigureRevisionsOperation {
	return &ConfigureRevisionsOperation{
		_configuration: configuration,
	}
}

// GetCommand returns new RavenCommand for this operation
func (o *ConfigureRevisionsOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewConfigureRevisionsCommand(o._configuration)
	return o.Command
}

var _ RavenCommand = &ConfigureRevisionsCommand{}

// ConfigureRevisionsCommand represents configure revisions command
type ConfigureRevisionsCommand struct {
	RavenCommandBase

	_configuration *RevisionsConfiguration

	Result *ConfigureRevisionsOperationResult
}

// NewConfigureRevisionsCommand returns new ConfigureRevisionsCommand
func NewConfigureRevisionsCommand(configuration *RevisionsConfiguration) *ConfigureRevisionsCommand {
	cmd := &ConfigureRevisionsCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_configuration: configuration,
	}
	return cmd
}

// CreateRequest creates http request for this command
func (c *ConfigureRevisionsCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/admin/revisions/config"

	d, err := jsonMarshal(c._configuration)
	if err != nil {
		return nil, err
	}
	return NewHttpPost(url, d)
}

// SetResponse sets response for this command
func (c *ConfigureRevisionsCommand) SetResponse(response []byte, fromCache bool) error {
	return jsonUnmarshal(response, &c.Result)
}

// ConfigureRevisionsOperationResult represents result of configure revisions operation
type ConfigureRevisionsOperationResult struct {
	RaftCommandIndex int64 `json:"RaftCommandIndex"`
}
