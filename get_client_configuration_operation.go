package ravendb

import (
	"encoding/json"
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

func (o *GetClientConfigurationOperation) getCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewGetClientConfigurationCommand()
	return o.Command
}

type GetClientConfigurationCommand struct {
	*RavenCommandBase

	Result *GetClientConfigurationCommandResult
}

func NewGetClientConfigurationCommand() *GetClientConfigurationCommand {
	cmd := &GetClientConfigurationCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}
	return cmd
}

func (c *GetClientConfigurationCommand) createRequest(node *ServerNode) (*http.Request, error) {

	url := node.getUrl() + "/databases/" + node.getDatabase() + "/configuration/client"

	return NewHttpGet(url)
}

func (c *GetClientConfigurationCommand) setResponse(response string, fromCache bool) error {
	if response == "" {
		return nil
	}

	var res GetClientConfigurationCommandResult
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}

type GetClientConfigurationCommandResult struct {
	Etag          int                  `json:"Etag"`
	Configuration *ClientConfiguration `json:"Configuration"`
}

func (r *GetClientConfigurationCommandResult) getEtag() int {
	return r.Etag
}

func (r *GetClientConfigurationCommandResult) setEtag(etag int) {
	r.Etag = etag
}

func (r *GetClientConfigurationCommandResult) getConfiguration() *ClientConfiguration {
	return r.Configuration
}

func (r *GetClientConfigurationCommandResult) setConfiguration(configuration *ClientConfiguration) {
	r.Configuration = configuration
}
