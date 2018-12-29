package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func deleteTestDeleteDocumentByEntity(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	newSession := openSessionMust(t, store)

	{
		user := &User{}
		user.setName("RavenDB")

		err = newSession.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)
	}

	var user *User
	err = newSession.Load(&user, "users/1")
	assert.NoError(t, err)
	assert.NotNil(t, user)

	err = newSession.DeleteEntity(user)
	assert.NoError(t, err)
	err = newSession.SaveChanges()
	assert.NoError(t, err)

	var nilUser *User
	err = newSession.Load(&nilUser, "users/1")
	assert.NoError(t, err)
	assert.Nil(t, nilUser)
	newSession.Close()
}

func deleteTestDeleteDocumentById(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	newSession := openSessionMust(t, store)

	user := &User{}
	user.setName("RavenDB")

	err := newSession.StoreWithID(user, "users/1")
	assert.NoError(t, err)
	err = newSession.SaveChanges()
	assert.NoError(t, err)

	{
		var user *User
		err = newSession.Load(&user, "users/1")
		assert.NoError(t, err)
		assert.NotNil(t, user)
	}

	err = newSession.Delete("users/1")
	assert.NoError(t, err)
	err = newSession.SaveChanges()
	assert.NoError(t, err)

	var nilUser *User
	err = newSession.Load(&nilUser, "users/1")
	assert.NoError(t, err)
	assert.Nil(t, nilUser)
	newSession.Close()
}

func TestDelete(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	deleteTestDeleteDocumentByEntity(t, driver)
	deleteTestDeleteDocumentById(t, driver)
}
