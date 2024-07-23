package ravendb

import (
	"encoding/json"
	"net/http"
)

var _ IVoidMaintenanceOperation = &ConfigureExpirationOperation{}

type ConfigureExpirationOperation struct {
	parameters *ExpirationConfiguration
	Command    *ConfigureExpirationCommand
}

func NewConfigureExpirationOperationWithConfiguration(expirationConfiguration *ExpirationConfiguration) (*ConfigureExpirationOperation, error) {
	return &ConfigureExpirationOperation{
		parameters: expirationConfiguration,
	}, nil
}

func NewConfigureExpirationOperation(disabled bool, deleteFrequencyInSec *int64, maxItemsToProcess *int64) (*ConfigureExpirationOperation, error) {

	p := &ExpirationConfiguration{
		Disabled:             disabled,
		DeleteFrequencyInSec: deleteFrequencyInSec,
		MaxItemsToProcess:    maxItemsToProcess,
	}
	return &ConfigureExpirationOperation{
		parameters: p,
	}, nil
}

type ConfigureExpirationCommand struct {
	RavenCommandBase
	_parameters []byte
	Result      *ExpirationConfigurationResult
}

// GetCommand returns a command
func (o *ConfigureExpirationOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	var err error
	o.Command, err = newConfigureExpirationCommand(conventions, o.parameters)
	if err != nil {
		return nil, err
	}
	return o.Command, nil
}

func newConfigureExpirationCommand(conventions *DocumentConventions, parameters *ExpirationConfiguration) (*ConfigureExpirationCommand, error) {
	if conventions == nil {
		return nil, newIllegalArgumentError("conventions cannot be null")
	}
	if parameters == nil {
		return nil, newIllegalArgumentError("parameters cannot be null")
	}

	// Note: compared to Java, we shortcut things by serializing to JSON
	// here as it's simpler and faster than two-step serialization,
	// first to map[string]interface{} and then to JSON
	d, err := jsonMarshal(parameters)
	panicIf(err != nil, "jsonMarshal failed with %s", err)
	cmd := &ConfigureExpirationCommand{
		RavenCommandBase: NewRavenCommandBase(),
		_parameters:      d,
	}
	cmd.ResponseType = RavenCommandResponseTypeObject
	return cmd, nil
}

func (c *ConfigureExpirationCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/admin/expiration/config"

	return NewHttpPost(url, c._parameters)
}

func (c *ConfigureExpirationCommand) SetResponse(response []byte, fromCache bool) error {
	return json.Unmarshal(response, &c.Result)
}

// ExpirationConfiguration
type ExpirationConfiguration struct {
	Disabled             bool   `json:"Disabled"`
	DeleteFrequencyInSec *int64 `json:"DeleteFrequencyInSec"`
	MaxItemsToProcess    *int64 `json:"MaxItemsToProcess"`
}

type ExpirationConfigurationResult struct {
	RaftCommandIndex *int64 `json:"RaftCommandIndex"`
}
