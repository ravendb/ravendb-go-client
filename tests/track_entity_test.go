package tests

import (
	"strings"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func trackEntityTest_deletingEntityThatIsNotTrackedShouldThrow(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		err = session.DeleteEntity(&User{})
		assert.Error(t, err)
		_ = err.(*ravendb.IllegalStateException)
		msg := err.Error()
		assert.True(t, strings.HasSuffix(msg, "is not associated with the session, cannot delete unknown entity instance"))
		session.Close()
	}
}

func trackEntityTest_loadingDeletedDocumentShouldReturnNull(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user1 := &User{}
		user1.setName("John")
		user1.ID = "users/1"

		user2 := &User{}
		user2.setName("Jonathan")
		user2.ID = "users/2"

		err = session.Store(user1)
		assert.NoError(t, err)
		err = session.Store(user2)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		session := openSessionMust(t, store)
		err = session.Delete("users/1")
		assert.NoError(t, err)
		err = session.Delete("users/2")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		session := openSessionMust(t, store)
		_, err = session.LoadOld(ravendb.GetTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		_, err = session.LoadOld(ravendb.GetTypeOf(&User{}), "users/2")
		assert.NoError(t, err)
		session.Close()
	}
}

func trackEntityTest_storingDocumentWithTheSameIdInTheSameSessionShouldThrow(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user := &User{}
		user.ID = "users/1"
		user.setName("User1")

		err = session.Store(user)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		newUser := &User{}
		newUser.setName("User2")
		newUser.ID = "users/1"

		err = session.Store(newUser)
		_ = err.(*ravendb.NonUniqueObjectException)
		msg := err.Error()
		assert.True(t, strings.HasPrefix(msg, "Attempted to associate a different object with id 'users/1'"))
		session.Close()
	}
}

func TestTrackEntity(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of java tests
	trackEntityTest_loadingDeletedDocumentShouldReturnNull(t)
	trackEntityTest_deletingEntityThatIsNotTrackedShouldThrow(t)
	trackEntityTest_storingDocumentWithTheSameIdInTheSameSessionShouldThrow(t)
}
