package ravendb

import (
	"encoding/json"
	"net/http"
)

var _ IVoidMaintenanceOperation = &SetIndexesPriorityOperation{}

type SetIndexesPriorityOperation struct {
	_parameters *SetIndexesPriorityParameters
	Command     *SetIndexesPriorityCommand
}

func NewSetIndexesPriorityOperation(indexName string, priority IndexPriority) *SetIndexesPriorityOperation {
	panicIf(indexName == "", "indexName cannot be empty")

	p := &SetIndexesPriorityParameters{
		IndexNames: []string{indexName},
		Priority:   priority,
	}
	return NewSetIndexesPriorityOperationWithParameters(p)
}

func NewSetIndexesPriorityOperationWithParameters(parameters *SetIndexesPriorityParameters) *SetIndexesPriorityOperation {
	panicIf(parameters == nil, "parameters cannot be nil")

	return &SetIndexesPriorityOperation{
		_parameters: parameters,
	}
}

func (o *SetIndexesPriorityOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewSetIndexesPriorityCommand(conventions, o._parameters)
	return o.Command
}

var (
	_ RavenCommand = &SetIndexesPriorityCommand{}
)

type SetIndexesPriorityCommand struct {
	*RavenCommandBase

	_parameters []byte
}

func NewSetIndexesPriorityCommand(conventions *DocumentConventions, parameters *SetIndexesPriorityParameters) *SetIndexesPriorityCommand {
	panicIf(conventions == nil, "conventions cannot be null")
	panicIf(parameters == nil, "parameters cannot be null")

	// Note: compared to Java, we shortcut things by serializing to JSON
	// here as it's simpler and faster than two-step serialization,
	// first to ObjectNode and then to JSON
	d, err := json.Marshal(parameters)
	panicIf(err != nil, "json.Marshal failed with %s", err)
	cmd := &SetIndexesPriorityCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_parameters: d,
	}
	cmd.ResponseType = RavenCommandResponseType_EMPTY
	return cmd
}

func (c *SetIndexesPriorityCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/indexes/set-priority"

	return NewHttpPost(url, c._parameters)
}

// Note: in Java it's Parameters class nested in SetIndexesPriorityOperation
// Parameters is already taken
type SetIndexesPriorityParameters struct {
	IndexNames []string      `json:"IndexNames"`
	Priority   IndexPriority `json:"Priority"`
}

func (p *SetIndexesPriorityParameters) getIndexNames() []string {
	return p.IndexNames
}

func (p *SetIndexesPriorityParameters) setIndexNames(indexNames []string) {
	p.IndexNames = indexNames
}

func (p *SetIndexesPriorityParameters) getPriority() IndexLockMode {
	return p.Priority
}

func (p *SetIndexesPriorityParameters) setPriority(priority IndexLockMode) {
	p.Priority = priority
}
