package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

func deleteTest_deleteDocumentByEntity(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	newSession := openSessionMust(t, store)

	user := &User{}
	user.setName("RavenDB")

	err := newSession.StoreWithID(user, "users/1")
	assert.NoError(t, err)
	err = newSession.SaveChanges()
	assert.NoError(t, err)

	result, err := newSession.Load(ravendb.GetTypeOf(&User{}), "users/1")
	assert.NoError(t, err)
	user = result.(*User)

	assert.NotNil(t, user)

	err = newSession.DeleteEntity(user)
	assert.NoError(t, err)
	err = newSession.SaveChanges()
	assert.NoError(t, err)

	result, err = newSession.Load(ravendb.GetTypeOf(&User{}), "users/1")
	assert.NoError(t, err)
	nilUser := result.(*User)
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

	result, err := newSession.Load(ravendb.GetTypeOf(&User{}), "users/1")
	assert.NoError(t, err)
	user = result.(*User)
	assert.NotNil(t, user)

	err = newSession.Delete("users/1")
	assert.NoError(t, err)
	err = newSession.SaveChanges()
	assert.NoError(t, err)

	result, err = newSession.Load(ravendb.GetTypeOf(&User{}), "users/1")
	assert.NoError(t, err)
	nilUser := result.(*User)
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
