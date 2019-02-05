package ravendb

import (
	"net/http"
)

var (
	// make sure GetClientConfigurationOperation implements IMaintenanceOperation
	_ IMaintenanceOperation = &GetClientConfigurationOperation{}
)

type GetClientConfigurationOperation struct {
	Command *GetClientConfigurationCommand
}

func NewGetClientConfigurationOperation() *GetClientConfigurationOperation {
	return &GetClientConfigurationOperation{}
}

func (o *GetClientConfigurationOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	o.Command = NewGetClientConfigurationCommand()
	return o.Command, nil
}

type GetClientConfigurationCommand struct {
	RavenCommandBase

	Result *GetClientConfigurationCommandResult
}

func NewGetClientConfigurationCommand() *GetClientConfigurationCommand {
	cmd := &GetClientConfigurationCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}
	return cmd
}

func (c *GetClientConfigurationCommand) createRequest(node *ServerNode) (*http.Request, error) {

	url := node.URL + "/databases/" + node.Database + "/configuration/client"

	return newHttpGet(url)
}

func (c *GetClientConfigurationCommand) setResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return nil
	}

	return jsonUnmarshal(response, &c.Result)
}

type GetClientConfigurationCommandResult struct {
	Etag          int64                `json:"Etag"`
	Configuration *ClientConfiguration `json:"Configuration"`
}
