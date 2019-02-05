package ravendb

import (
	"net/http"
)

var (
	_ IMaintenanceOperation = &UpdateExternalReplicationOperation{}
)

type UpdateExternalReplicationOperation struct {
	_newWatcher *ExternalReplication

	Command *UpdateExternalReplicationCommand
}

func NewUpdateExternalReplicationOperation(newWatcher *ExternalReplication) *UpdateExternalReplicationOperation {
	return &UpdateExternalReplicationOperation{
		_newWatcher: newWatcher,
	}
}

func (o *UpdateExternalReplicationOperation) GetCommand(conventions *DocumentConventions) (RavenCommand, error) {
	o.Command = NewUpdateExternalReplicationCommand(o._newWatcher)
	return o.Command, nil
}

var _ RavenCommand = &UpdateExternalReplicationCommand{}

type UpdateExternalReplicationCommand struct {
	RavenCommandBase

	_newWatcher *ExternalReplication

	Result *ModifyOngoingTaskResult
}

func NewUpdateExternalReplicationCommand(newWatcher *ExternalReplication) *UpdateExternalReplicationCommand {
	cmd := &UpdateExternalReplicationCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_newWatcher: newWatcher,
	}
	return cmd
}

func (c *UpdateExternalReplicationCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/admin/tasks/external-replication"

	m := map[string]interface{}{
		"Watcher": c._newWatcher,
	}
	d, err := jsonMarshal(m)
	if err != nil {
		return nil, err
	}
	return NewHttpPost(url, d)
}

func (c *UpdateExternalReplicationCommand) SetResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}

	return jsonUnmarshal(response, &c.Result)
}
