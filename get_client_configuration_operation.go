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
}

func NewGetClientConfigurationOperation() *GetClientConfigurationOperation {
	return &GetClientConfigurationOperation{}
}

func (o *GetClientConfigurationOperation) getCommand(conventions *DocumentConventions) *RavenCommand {
	return NewGetClientConfigurationCommand()
}

func NewGetClientConfigurationCommand() *RavenCommand {
	cmd := NewRavenCommand()
	cmd.createRequestFunc = GetClientConfigurationOperation_createRequest
	cmd.setResponseFunc = GetClientConfigurationOperation_setResponse

	return cmd
}

func GetClientConfigurationOperation_createRequest(cmd *RavenCommand, node *ServerNode) (*http.Request, error) {

	url := node.getUrl() + "/databases/" + node.getDatabase() + "/configuration/client"

	return NewHttpGet(url)
}

func GetClientConfigurationOperation_setResponse(cmd *RavenCommand, response String, fromCache bool) error {
	if response == "" {
		return nil
	}

	var res GetClientConfigurationCommandResult
	err := json.Unmarshal([]byte(response), &res)
	if err != nil {
		return err
	}
	cmd.result = &res
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
