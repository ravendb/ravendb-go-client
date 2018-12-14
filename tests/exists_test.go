package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func existsTest_checkIfDocumentExists(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
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

		var user *User
		err = session.Load(&user, "users/2")
		assert.NoError(t, err)
		assert.NotNil(t, user)
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

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	existsTest_checkIfDocumentExists(t, driver)
}
