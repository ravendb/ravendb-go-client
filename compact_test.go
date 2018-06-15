package ravendb

import (
	"strings"
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

func compactTest_canCompactDatabase(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	{
		newSession := openSessionMust(t, store)
		user1 := NewUser()
		user1.setLastName("user1")
		err = newSession.StoreEntityWithID(user1, "users/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)
	}

	compactSettings := NewCompactSettings()
	compactSettings.setDatabaseName(store.getDatabase())
	compactSettings.setDocuments(true)

	compactOp := NewCompactDatabaseOperation(compactSettings)
	operation, err := store.maintenance().server().sendAsync(compactOp)
	assert.NoError(t, err)

	// we can't compact in memory database but here we just test is request was send successfully
	err = operation.waitForCompletion()
	msg := err.Error()
	assert.True(t, strings.Contains(msg, "Unable to cast object of type 'PureMemoryStorageEnvironmentOptions' to type 'DirectoryStorageEnvironmentOptions'"))
}

func TestCompact(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	if useProxy() {
		proxy.ChangeLogFile("trace_compact_go.txt")
	}

	createTestDriver()
	defer deleteTestDriver()

	// follows execution order of java tests
	compactTest_canCompactDatabase(t)
}
