package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

func existsTest_checkIfDocumentExists(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		assert.NoError(t, err)
		idan := &User{}
		idan.setName("Idan")

		shalom := &User{}
		shalom.setName("Shalom")

		err = session.StoreWithID(idan, "users/1")
		assert.NoError(t, err)
		err = session.StoreWithID(shalom, "users/2")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		session := openSessionMust(t, store)
		assert.NoError(t, err)
		ok, err := session.Advanced().Exists("users/1")
		assert.NoError(t, err)
		assert.True(t, ok)

		ok, err = session.Advanced().Exists("users/10")
		assert.NoError(t, err)
		assert.False(t, ok)

		_, err = session.LoadOld(ravendb.GetTypeOf(&User{}), "users/2")
		assert.NoError(t, err)
		ok, err = session.Advanced().Exists("users/2")
		assert.NoError(t, err)
		assert.True(t, ok)
		session.Close()
	}
}

func TestExists(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	existsTest_checkIfDocumentExists(t)
}
