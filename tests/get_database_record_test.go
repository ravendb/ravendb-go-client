package tests

import (
	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
	"testing"
)

func getDatabaseRecordCanGetDatabaseRecord(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	op := ravendb.NewGetDatabaseRecordOperation(store.GetDatabase())
	err = store.Maintenance().Server().Send(op)
	assert.NoError(t, err)
	assert.Equal(t, op.Command.Result.DatabaseName, store.GetDatabase())
}

func TestGetDatabaseRecord(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	getDatabaseRecordCanGetDatabaseRecord(t, driver)
}
