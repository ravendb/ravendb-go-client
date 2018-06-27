package ravendb

import (
	"encoding/json"
	"net/http"
)

var _ IVoidMaintenanceOperation = &SetIndexesLockOperation{}

type SetIndexesLockOperation struct {
	_parameters *SetIndexesLockParameters
	Command     *SetIndexesLockCommand
}

func NewSetIndexesLockOperation(indexName string, mode IndexLockMode) *SetIndexesLockOperation {
	panicIf(indexName == "", "indexName cannot be empty")

	p := &SetIndexesLockParameters{
		IndexNames: []string{indexName},
		Mode:       mode,
	}
	return NewSetIndexesLockOperationWithParameters(p)
}

func NewSetIndexesLockOperationWithParameters(parameters *SetIndexesLockParameters) *SetIndexesLockOperation {
	panicIf(parameters == nil, "parameters cannot be nil")

	return &SetIndexesLockOperation{
		_parameters: parameters,
	}
}

func (o *SetIndexesLockOperation) getCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewSetIndexesLockCommand(conventions, o._parameters)
	return o.Command
}

var (
	_ RavenCommand = &SetIndexesLockCommand{}
)

type SetIndexesLockCommand struct {
	*RavenCommandBase

	_parameters []byte
}

func NewSetIndexesLockCommand(conventions *DocumentConventions, parameters *SetIndexesLockParameters) *SetIndexesLockCommand {
	panicIf(conventions == nil, "conventions cannot be null")
	panicIf(parameters == nil, "parameters cannot be null")

	// Note: compared to Java, we shortcut things by serializing to JSON
	// here as it's simpler and faster than two-step serialization,
	// first to ObjectNode and then to JSON
	d, err := json.Marshal(parameters)
	panicIf(err != nil, "json.Marshal failed with %s", err)
	cmd := &SetIndexesLockCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_parameters: d,
	}
	cmd.responseType = RavenCommandResponseType_EMPTY
	return cmd
}

func (c *SetIndexesLockCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/indexes/set-lock"

	return NewHttpPost(url, c._parameters)
}

// Note: in Java it's Parameters class nested in SetIndexesLockOperation
// Parameters is already taken
type SetIndexesLockParameters struct {
	IndexNames []string      `json:"IndexNames"`
	Mode       IndexLockMode `json:"Mode"`
}

func (p *SetIndexesLockParameters) getIndexNames() []string {
	return p.IndexNames
}

func (p *SetIndexesLockParameters) setIndexNames(indexNames []string) {
	p.IndexNames = indexNames
}

func (p *SetIndexesLockParameters) getMode() IndexLockMode {
	return p.Mode
}

func (p *SetIndexesLockParameters) setMode(mode IndexLockMode) {
	p.Mode = mode
}
