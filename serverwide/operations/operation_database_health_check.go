package operations

import (
	"github.com/ravendb/ravendb-go-client"
	"net/http"
)

type OperationDatabaseHealthCheck struct {}

func (operation * OperationDatabaseHealthCheck) GetCommand(conventions *ravendb.DocumentConventions) (ravendb.RavenCommand, error) {
	return &getDatabaseHealthCheckOperation{
		RavenCommandBase: ravendb.RavenCommandBase{
			ResponseType: ravendb.RavenCommandResponseTypeEmpty,
		},
	}, nil
}

type getDatabaseHealthCheckOperation struct {
	ravendb.RavenCommandBase
}

func (c *getDatabaseHealthCheckOperation) CreateRequest(node *ravendb.ServerNode) (*http.Request, error) {
	url := node.URL + "/databases/" + node.Database + "/healthcheck"
	return http.NewRequest(http.MethodGet, url, nil)
}
func (c *getDatabaseHealthCheckOperation) SetResponse(response []byte, fromCache bool) error {
	return nil
}
