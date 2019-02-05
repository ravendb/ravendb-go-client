package ravendb

import (
	"net/http"
	"strings"
)

var _ IVoidMaintenanceOperation = &SetIndexesLockOperation{}

type SetIndexesLockOperation struct {
	parameters *SetIndexesLockParameters
	Command    *SetIndexesLockCommand
}

func NewSetIndexesLockOperation(indexName string, mode IndexLockMode) (*SetIndexesLockOperation, error) {
	if indexName == "" {
		return nil, newIllegalArgumentError("indexName cannot be empty")
	}

	p := &SetIndexesLockParameters{
		IndexNames: []string{indexName},
		Mode:       mode,
	}
	return NewSetIndexesLockOperationWithParameters(p)
}

func NewSetIndexesLockOperationWithParameters(parameters *SetIndexesLockParameters) (*SetIndexesLockOperation, error) {
	if parameters == nil {
		return nil, newIllegalArgumentError("parameters cannot be nil")
	}

	res := &SetIndexesLockOperation{
		parameters: parameters,
	}
	if err := res.filterAutoIndexes(); err != nil {
		return nil, err
	}
	return res, nil
}

func (o *SetIndexesLockOperation) filterAutoIndexes() error {
	// Check for auto-indexes - we do not set lock for auto-indexes
	for _, indexName := range o.parameters.IndexNames {
		s := strings.ToLower(indexName)
		if strings.HasPrefix(s, "auto/") {
			return newIllegalArgumentError("Indexes list contains Auto-Indexes. Lock Mode is not set for Auto-Indexes.")
		}
	}
	return nil
}

func (o *SetIndexesLockOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	var err error
	o.Command, err = NewSetIndexesLockCommand(conventions, o.parameters)
	if err != nil {
		return nil, err
	}
	return o.Command, nil
}

var (
	_ RavenCommand = &SetIndexesLockCommand{}
)

type SetIndexesLockCommand struct {
	RavenCommandBase

	parameters []byte
}

func NewSetIndexesLockCommand(conventions *DocumentConventions, parameters *SetIndexesLockParameters) (*SetIndexesLockCommand, error) {
	if conventions == nil {
		return nil, newIllegalArgumentError("conventions cannot be null")
	}
	if parameters == nil {
		return nil, newIllegalArgumentError("parameters cannot be null")
	}

	// Note: compared to Java, we shortcut things by serializing to JSON
	// here as it's simpler and faster than two-step serialization,
	// first to ObjectNode and then to JSON
	d, err := jsonMarshal(parameters)
	if err != nil {
		return nil, err
	}
	cmd := &SetIndexesLockCommand{
		RavenCommandBase: NewRavenCommandBase(),

		parameters: d,
	}
	cmd.ResponseType = RavenCommandResponseTypeEmpty
	return cmd, nil
}

func (c *SetIndexesLockCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/indexes/set-lock"

	return newHttpPost(url, c.parameters)
}

// Note: in Java it's Parameters class nested in SetIndexesLockOperation
// Parameters is already taken
type SetIndexesLockParameters struct {
	IndexNames []string      `json:"IndexNames"`
	Mode       IndexLockMode `json:"Mode"`
}
