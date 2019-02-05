package ravendb

import (
	"net/http"
)

var (
	_ IMaintenanceOperation = &ConfigureRevisionsOperation{}
)

// ConfigureRevisionsOperation represents configure revisions operation
type ConfigureRevisionsOperation struct {
	configuration *RevisionsConfiguration
	Command       *ConfigureRevisionsCommand
}

// NewConfigureRevisionsOperation returns new ConfigureRevisionsOperation
func NewConfigureRevisionsOperation(configuration *RevisionsConfiguration) *ConfigureRevisionsOperation {
	return &ConfigureRevisionsOperation{
		configuration: configuration,
	}
}

// GetCommand returns new RavenCommand for this operation
func (o *ConfigureRevisionsOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	o.Command = NewConfigureRevisionsCommand(o.configuration)
	return o.Command, nil
}

var _ RavenCommand = &ConfigureRevisionsCommand{}

// ConfigureRevisionsCommand represents configure revisions command
type ConfigureRevisionsCommand struct {
	RavenCommandBase

	configuration *RevisionsConfiguration

	Result *ConfigureRevisionsOperationResult
}

// NewConfigureRevisionsCommand returns new ConfigureRevisionsCommand
func NewConfigureRevisionsCommand(configuration *RevisionsConfiguration) *ConfigureRevisionsCommand {
	cmd := &ConfigureRevisionsCommand{
		RavenCommandBase: NewRavenCommandBase(),

		configuration: configuration,
	}
	return cmd
}

func (c *ConfigureRevisionsCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/admin/revisions/config"

	d, err := jsonMarshal(c.configuration)
	if err != nil {
		return nil, err
	}
	return newHttpPost(url, d)
}

func (c *ConfigureRevisionsCommand) setResponse(response []byte, fromCache bool) error {
	return jsonUnmarshal(response, &c.Result)
}

// ConfigureRevisionsOperationResult represents result of configure revisions operation
type ConfigureRevisionsOperationResult struct {
	RaftCommandIndex int64 `json:"RaftCommandIndex"`
}
