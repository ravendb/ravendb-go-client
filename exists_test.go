package ravendb

import (
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

func existsTest_checkIfDocumentExists(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	{
		session := openSessionMust(t, store)
		assert.NoError(t, err)
		idan := NewUser()
		idan.setName("Idan")

		shalom := NewUser()
		shalom.setName("Shalom")

		err = session.StoreEntityWithID(idan, "users/1")
		assert.NoError(t, err)
		err = session.StoreEntityWithID(shalom, "users/2")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
	}

	{
		session := openSessionMust(t, store)
		assert.NoError(t, err)
		ok, err := session.advanced().exists("users/1")
		assert.NoError(t, err)
		assert.True(t, ok)

		ok, err = session.advanced().exists("users/10")
		assert.NoError(t, err)
		assert.False(t, ok)

		_, err = session.load(getTypeOfValue(NewUser()), "users/2")
		assert.NoError(t, err)
		ok, err = session.advanced().exists("users/2")
		assert.NoError(t, err)
		assert.True(t, ok)
	}
}

func TestExists(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	if useProxy() {
		proxy.ChangeLogFile("trace_exists_go.txt")
	}

	createTestDriver()
	defer deleteTestDriver()

	existsTest_checkIfDocumentExists(t)
}
