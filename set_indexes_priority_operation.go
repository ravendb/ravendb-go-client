package ravendb

import (
	"net/http"
)

var _ IVoidMaintenanceOperation = &SetIndexesPriorityOperation{}

// SetIndexesPriorityOperation represents operation for setting indexes priority
type SetIndexesPriorityOperation struct {
	_parameters *SetIndexesPriorityParameters
	Command     *SetIndexesPriorityCommand
}

// NewSetIndexesPriorityOperation returns new SetIndexesPriorityParameters
func NewSetIndexesPriorityOperation(indexName string, priority IndexPriority) *SetIndexesPriorityOperation {
	panicIf(indexName == "", "indexName cannot be empty")

	p := &SetIndexesPriorityParameters{
		IndexNames: []string{indexName},
		Priority:   priority,
	}
	return &SetIndexesPriorityOperation{
		_parameters: p,
	}
}

// GetCommand returns a command
func (o *SetIndexesPriorityOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewSetIndexesPriorityCommand(conventions, o._parameters)
	return o.Command
}

var (
	_ RavenCommand = &SetIndexesPriorityCommand{}
)

// SetIndexesPriorityCommand represents command to set indexes priority
type SetIndexesPriorityCommand struct {
	RavenCommandBase

	_parameters []byte
}

// NewSetIndexesPriorityCommand returns new SetIndexesPriorityCommand
func NewSetIndexesPriorityCommand(conventions *DocumentConventions, parameters *SetIndexesPriorityParameters) *SetIndexesPriorityCommand {
	panicIf(conventions == nil, "conventions cannot be null")
	panicIf(parameters == nil, "parameters cannot be null")

	// Note: compared to Java, we shortcut things by serializing to JSON
	// here as it's simpler and faster than two-step serialization,
	// first to map[string]interface{} and then to JSON
	d, err := jsonMarshal(parameters)
	panicIf(err != nil, "jsonMarshal failed with %s", err)
	cmd := &SetIndexesPriorityCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_parameters: d,
	}
	cmd.ResponseType = RavenCommandResponseTypeEmpty
	return cmd
}

// CreateRequest creates http request for the command
func (c *SetIndexesPriorityCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/indexes/set-priority"

	return NewHttpPost(url, c._parameters)
}

// SetIndexesPriorityParameters represents arrgument for SetIndexPriorityCommand
// Note: in Java it's Parameters class nested in SetIndexesPriorityOperation
// "Parameters" name is already taken
type SetIndexesPriorityParameters struct {
	IndexNames []string      `json:"IndexNames"`
	Priority   IndexPriority `json:"Priority"`
}
