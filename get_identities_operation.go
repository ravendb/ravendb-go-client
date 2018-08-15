package ravendb

import (
	"encoding/json"
	"net/http"
)

var (
	_ IMaintenanceOperation = &GetIdentitiesOperation{}
)

type GetIdentitiesOperation struct {
	Command *GetIdentitiesCommand
}

func NewGetIdentitiesOperation() *GetIdentitiesOperation {
	return &GetIdentitiesOperation{}
}

func (o *GetIdentitiesOperation) GetCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewGetIdentitiesCommand()
	return o.Command
}

type GetIdentitiesCommand struct {
	*RavenCommandBase

	Result map[string]int
}

func NewGetIdentitiesCommand() *GetIdentitiesCommand {
	cmd := &GetIdentitiesCommand{
		RavenCommandBase: NewRavenCommandBase(),
	}
	cmd.IsReadRequest = true
	return cmd
}

func (c *GetIdentitiesCommand) CreateRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/debug/identities"

	return NewHttpGet(url)

}

func (c *GetIdentitiesCommand) SetResponse(response []byte, fromCache bool) error {
	var res map[string]int
	err := json.Unmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = res
	return nil
}
