package ravendb

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func trackEntityTest_deletingEntityThatIsNotTrackedShouldThrow(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		err = session.DeleteEntity(NewUser())
		assert.Error(t, err)
		_ = err.(*IllegalStateException)
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
		user1 := NewUser()
		user1.setName("John")
		user1.setId("users/1")

		user2 := NewUser()
		user2.setName("Jonathan")
		user2.setId("users/2")

		err = session.StoreEntity(user1)
		assert.NoError(t, err)
		err = session.StoreEntity(user2)
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
		_, err = session.load(getTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		_, err = session.load(getTypeOf(&User{}), "users/2")
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
		user := NewUser()
		user.setId("users/1")
		user.setName("User1")

		err = session.StoreEntity(user)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		newUser := NewUser()
		newUser.setName("User2")
		newUser.setId("users/1")

		err = session.StoreEntity(newUser)
		_ = err.(*NonUniqueObjectException)
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
	defer func() {
		r := recover()
		destroyDriver()
		if r != nil {
			panic(r)
		}
	}()

	// matches order of java tests
	trackEntityTest_loadingDeletedDocumentShouldReturnNull(t)
	trackEntityTest_deletingEntityThatIsNotTrackedShouldThrow(t)
	trackEntityTest_storingDocumentWithTheSameIdInTheSameSessionShouldThrow(t)
}
