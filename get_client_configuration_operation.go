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

func (o *GetClientConfigurationOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
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

func (c *GetClientConfigurationCommand) CreateRequest(node *ServerNode) (*http.Request, error) {

	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/configuration/client"

	return NewHttpGet(url)
}

func (c *GetClientConfigurationCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return nil
	}

	var res GetClientConfigurationCommandResult
	err := json.Unmarshal(response, &res)
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

func (r *GetClientConfigurationCommandResult) GetEtag() int {
	return r.Etag
}

func (r *GetClientConfigurationCommandResult) SetEtag(etag int) {
	r.Etag = etag
}

func (r *GetClientConfigurationCommandResult) GetConfiguration() *ClientConfiguration {
	return r.Configuration
}

func (r *GetClientConfigurationCommandResult) SetConfiguration(configuration *ClientConfiguration) {
	r.Configuration = configuration
}
