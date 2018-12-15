package ravendb

import (
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
	RavenCommandBase

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
	url := node.GetUrl() + "/databases/" + node.GetDatabase() + "/debug/identities"

	return NewHttpGet(url)

}

func (c *GetIdentitiesCommand) SetResponse(response []byte, fromCache bool) error {
	return jsonUnmarshal(response, &c.Result)
}
