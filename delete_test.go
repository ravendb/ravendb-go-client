package ravendb

import (
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

func deleteTest_deleteDocumentByEntity(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	newSession := openSessionMust(t, store)

	user := NewUser()
	user.setName("RavenDB")

	err := newSession.StoreEntityWithID(user, "users/1")
	assert.NoError(t, err)
	err = newSession.SaveChanges()
	assert.NoError(t, err)

	result, err := newSession.load(getTypeOf(NewUser()), "users/1")
	assert.NoError(t, err)
	user = result.(*User)

	assert.NotNil(t, user)

	err = newSession.DeleteEntity(user)
	assert.NoError(t, err)
	err = newSession.SaveChanges()
	assert.NoError(t, err)

	result, err = newSession.load(getTypeOf(NewUser()), "users/1")
	assert.NoError(t, err)
	nilUser := result.(*User)
	assert.Nil(t, nilUser)
}

func deleteTest_deleteDocumentById(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	newSession := openSessionMust(t, store)

	user := NewUser()
	user.setName("RavenDB")

	err := newSession.StoreEntityWithID(user, "users/1")
	assert.NoError(t, err)
	err = newSession.SaveChanges()
	assert.NoError(t, err)

	result, err := newSession.load(getTypeOf(NewUser()), "users/1")
	assert.NoError(t, err)
	user = result.(*User)
	assert.NotNil(t, user)

	err = newSession.Delete("users/1")
	assert.NoError(t, err)
	err = newSession.SaveChanges()
	assert.NoError(t, err)

	result, err = newSession.load(getTypeOf(NewUser()), "users/1")
	assert.NoError(t, err)
	nilUser := result.(*User)
	assert.Nil(t, nilUser)
}

func TestDelete(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	if useProxy() {
		proxy.ChangeLogFile("trace_delete_go.txt")
	}

	createTestDriver()
	defer deleteTestDriver()

	// matches order of Java tests
	deleteTest_deleteDocumentByEntity(t)
	deleteTest_deleteDocumentById(t)
}
