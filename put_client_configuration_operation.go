package ravendb

import (
	"net/http"
)

var _ IVoidMaintenanceOperation = &PutClientConfigurationOperation{}

type PutClientConfigurationOperation struct {
	configuration *ClientConfiguration
	Command       *PutClientConfigurationCommand
}

func NewPutClientConfigurationOperation(configuration *ClientConfiguration) (*PutClientConfigurationOperation, error) {
	if configuration == nil {
		return nil, newIllegalArgumentError("Configuration cannot be null")
	}

	return &PutClientConfigurationOperation{
		configuration: configuration,
	}, nil
}

func (o *PutClientConfigurationOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	var err error
	o.Command, err = NewPutClientConfigurationCommand(conventions, o.configuration)
	if err != nil {
		return nil, err
	}
	return o.Command, nil
}

var (
	_ RavenCommand = &PutClientConfigurationCommand{}
)

type PutClientConfigurationCommand struct {
	RavenCommandBase

	configuration []byte
}

func NewPutClientConfigurationCommand(conventions *DocumentConventions, configuration *ClientConfiguration) (*PutClientConfigurationCommand, error) {
	if conventions == nil {
		return nil, newIllegalArgumentError("conventions cannot be null")
	}
	if configuration == nil {
		return nil, newIllegalArgumentError("configuration cannot be null")
	}

	d, err := jsonMarshal(configuration)
	if err != nil {
		return nil, err
	}
	cmd := &PutClientConfigurationCommand{
		RavenCommandBase: NewRavenCommandBase(),

		configuration: d,
	}
	cmd.ResponseType = RavenCommandResponseTypeEmpty
	return cmd, nil
}

func (c *PutClientConfigurationCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/admin/configuration/client"
	return NewHttpPut(url, c.configuration)
}
