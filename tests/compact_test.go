package tests

import (
	"strings"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func compactTest_canCompactDatabase(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		user1 := &User{}
		user1.setLastName("user1")
		err = newSession.StoreWithID(user1, "users/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}

	compactSettings := ravendb.NewCompactSettings()
	compactSettings.SetDatabaseName(store.GetDatabase())
	compactSettings.SetDocuments(true)

	compactOp := ravendb.NewCompactDatabaseOperation(compactSettings)
	operation, err := store.Maintenance().Server().SendAsync(compactOp)
	assert.NoError(t, err)

	// we can't compact in memory database but here we just test is request was send successfully
	err = operation.WaitForCompletion()
	msg := err.Error()
	assert.True(t, strings.Contains(msg, "Unable to cast object of type 'PureMemoryStorageEnvironmentOptions' to type 'DirectoryStorageEnvironmentOptions'"))
}

func TestCompact(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// follows execution order of java tests
	compactTest_canCompactDatabase(t, driver)
}
