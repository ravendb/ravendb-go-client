package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func deleteTest_deleteDocumentByEntity(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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

	// TODO: should this be DeleteEntity(user)? Both?
	err = newSession.DeleteEntity(&user)
	assert.NoError(t, err)
	err = newSession.SaveChanges()
	assert.NoError(t, err)

	var nilUser *User
	err = newSession.Load(&nilUser, "users/1")
	assert.NoError(t, err)
	assert.Nil(t, nilUser)
	newSession.Close()
}

func deleteTest_deleteDocumentById(t *testing.T) {
	store := getDocumentStoreMust(t)
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
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	deleteTest_deleteDocumentByEntity(t)
	deleteTest_deleteDocumentById(t)
}
