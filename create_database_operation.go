package ravendb

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// ExecuteCreateDatabaseCommand executes CreateDatabaseCommand
func ExecuteCreateDatabaseCommand(exec CommandExecutorFunc, cmd *RavenCommand) (*DatabasePutResult, error) {
	var res DatabasePutResult
	err := excuteCmdAndJSONDecode(exec, cmd, &res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func NewCreateDatabaseOperation(databaseRecord *DatabaseRecord) *RavenCommand {
	return NewCreateDatabaseOperationWithReplicationFactor(databaseRecord, 1)
}

func NewCreateDatabaseOperationWithReplicationFactor(databaseRecord *DatabaseRecord, replicationFactor int) *RavenCommand {
	return NewCreateDatabaseCommand(databaseRecord, replicationFactor)
}

// NewCreateDatabaseCommand creates a new CreateDatabaseCommand
// TODO: Settings, SecureSettings
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/raven_operations/server_operations.py#L24
func NewCreateDatabaseCommand(databaseRecord *DatabaseRecord, replicationFactor int) *RavenCommand {
	panicIf(databaseRecord.DatabaseName == "", "DatabaseName empty in %#v", databaseRecord)
	databaseName := databaseRecord.DatabaseName
	if replicationFactor < 1 {
		replicationFactor = 1
	}
	url := fmt.Sprintf("{url}/admin/databases?name=%s&replication-factor=%d", databaseName, replicationFactor)
	data, err := json.Marshal(databaseRecord)
	must(err)
	res := &RavenCommand{
		Method:      http.MethodPut,
		URLTemplate: url,
		Data:        data,
	}
	return res
}
