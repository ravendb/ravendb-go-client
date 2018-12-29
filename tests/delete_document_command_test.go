package tests

import (
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func deleteDocumentCommandTestCanDeleteDocument(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user := &User{}
		user.setName("Marcin")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}
	command := ravendb.NewDeleteDocumentCommand("users/1", nil)
	err = store.GetRequestExecutor().ExecuteCommand(command)
	assert.NoError(t, err)
	{
		session := openSessionMust(t, store)
		var loadedUser *User
		err = session.Load(&loadedUser, "users/1")
		assert.NoError(t, err)
		assert.Nil(t, loadedUser)
		session.Close()
	}
}

func deleteDocumentCommandTestCanDeleteDocumentByEtag(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	var changeVector *string
	{
		session := openSessionMust(t, store)
		user := &User{}
		user.setName("Marcin")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		changeVector, err = session.Advanced().GetChangeVectorFor(user)
		assert.NoError(t, err)
		session.Close()
	}

	{
		session := openSessionMust(t, store)
		var loadedUser *User
		err = session.Load(&loadedUser, "users/1")
		assert.NoError(t, err)
		assert.NotNil(t, loadedUser)
		loadedUser.Age = 5
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	command := ravendb.NewDeleteDocumentCommand("users/1", changeVector)
	err = store.GetRequestExecutor().ExecuteCommand(command)
	assert.Error(t, err)
	_ = err.(*ravendb.ConcurrencyError)
}

func TestDeleteDocumentCommand(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// follows execution order of java tests
	deleteDocumentCommandTestCanDeleteDocument(t, driver)
	deleteDocumentCommandTestCanDeleteDocumentByEtag(t, driver)
}
