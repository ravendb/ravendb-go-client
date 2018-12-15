package ravendb

import (
	"net/http"
)

var (
	_ IMaintenanceOperation = &ConfigureRevisionsOperation{}
)

type ConfigureRevisionsOperation struct {
	_configuration *RevisionsConfiguration
	Command        *ConfigureRevisionsCommand
}

func NewConfigureRevisionsOperation(configuration *RevisionsConfiguration) *ConfigureRevisionsOperation {
	return &ConfigureRevisionsOperation{
		_configuration: configuration,
	}
}

func (o *ConfigureRevisionsOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewConfigureRevisionsCommand(conventions, o._configuration)
	return o.Command
}

var _ RavenCommand = &ConfigureRevisionsCommand{}

type ConfigureRevisionsCommand struct {
	RavenCommandBase

	_conventions   *DocumentConventions
	_configuration *RevisionsConfiguration

	Result *ConfigureRevisionsOperationResult
}

func NewConfigureRevisionsCommand(conventions *DocumentConventions, configuration *RevisionsConfiguration) *ConfigureRevisionsCommand {
	cmd := &ConfigureRevisionsCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_conventions:   conventions,
		_configuration: configuration,
	}
	return cmd
}

func (c *ConfigureRevisionsCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/admin/revisions/config"

	d, err := jsonMarshal(c._configuration)
	if err != nil {
		return nil, err
	}
	return NewHttpPost(url, d)
}

func (c *ConfigureRevisionsCommand) SetResponse(response []byte, fromCache bool) error {
	return jsonUnmarshal(response, &c.Result)
}

type ConfigureRevisionsOperationResult struct {
	Etag int `json:"ETag"`
}

func (r *ConfigureRevisionsOperationResult) getEtag() int {
	return r.Etag
}

func (r *ConfigureRevisionsOperationResult) setEtag(etag int) {
	r.Etag = etag
}
