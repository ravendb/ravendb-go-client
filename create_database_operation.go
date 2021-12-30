package ravendb

import (
	"net/http"
	"net/url"
	"strconv"
)

var _ IServerOperation = &CreateDatabaseOperation{}

// CreateDatabaseOperation represents "create database" operation
type CreateDatabaseOperation struct {
	databaseRecord    *DatabaseRecord
	replicationFactor int
}

// NewCreateDatabaseOperation returns CreateDatabaseOperation
func NewCreateDatabaseOperation(databaseRecord *DatabaseRecord, replicationFactor int) *CreateDatabaseOperation {
	if databaseRecord.DatabaseTopology != nil && databaseRecord.DatabaseTopology.ReplicationFactor > 0 {
		replicationFactor = databaseRecord.DatabaseTopology.ReplicationFactor
	} else if replicationFactor <= 0 {
		replicationFactor = 1
	}

	return &CreateDatabaseOperation{
		databaseRecord:    databaseRecord,
		replicationFactor: replicationFactor,
	}
}

// GetCommand returns command for this operation
func (o *CreateDatabaseOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	return NewCreateDatabaseCommand(conventions, o.databaseRecord, o.replicationFactor)
}

var (
	_ RavenCommand = &CreateDatabaseCommand{}
)

// CreateDatabaseCommand represents "create database" command
type CreateDatabaseCommand struct {
	RavenCommandBase

	conventions       *DocumentConventions
	databaseRecord    *DatabaseRecord
	replicationFactor int
	databaseName      string

	Result *DatabasePutResult
}

// NewCreateDatabaseCommand returns new CreateDatabaseCommand
func NewCreateDatabaseCommand(conventions *DocumentConventions, databaseRecord *DatabaseRecord, replicationFactor int) (*CreateDatabaseCommand, error) {
	if databaseRecord.DatabaseName == "" {
		return nil, newIllegalArgumentError("databaseRecord.DatabaseName cannot be empty")
	}
	if databaseRecord.DatabaseTopology != nil && databaseRecord.DatabaseTopology.ReplicationFactor > 0 {
		replicationFactor = databaseRecord.DatabaseTopology.ReplicationFactor
	} else if replicationFactor <= 0 {
		replicationFactor = 1
	}

	cmd := &CreateDatabaseCommand{
		RavenCommandBase: NewRavenCommandBase(),

		conventions:       conventions,
		databaseRecord:    databaseRecord,
		replicationFactor: replicationFactor,
		databaseName:      databaseRecord.DatabaseName,
	}
	return cmd, nil
}

func (c *CreateDatabaseCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/admin/databases?name=" + url.QueryEscape(c.databaseName)
	url += "&replicationFactor=" + strconv.Itoa(c.replicationFactor)

	js, err := jsonMarshal(c.databaseRecord)
	if err != nil {
		return nil, err
	}
	return newHttpPut(url, js)
}

func (c *CreateDatabaseCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}
	return jsonUnmarshal(response, &c.Result)
}
