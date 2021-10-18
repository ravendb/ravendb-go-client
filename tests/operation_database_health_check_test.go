package tests

import (
	"github.com/ravendb/ravendb-go-client"
	"github.com/ravendb/ravendb-go-client/serverwide/operations"
	"github.com/stretchr/testify/assert"
	"testing"
)

func getDatabaseHealthCheckTest(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()
	operation := operations.OperationDatabaseHealthCheck{}
	err = store.Maintenance().Send(&operation)
	assert.NoError(t, err)

	secondOperation := operations.OperationDatabaseHealthCheck{}
	err = store.Maintenance().ForDatabase("does_not_exists").Send(&secondOperation)
	_, ok := err.(*ravendb.DatabaseDoesNotExistError)
	assert.True(t, ok)
	assert.EqualError(t, err, err.(*ravendb.DatabaseDoesNotExistError).ErrorStr)
}

func TestGetDatabaseHealthCheckTest(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	getDatabaseHealthCheckTest(t, driver)
}
