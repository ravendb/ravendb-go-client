package tests

import (
	"reflect"
	"strings"
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func patchTestcanPatchSingleDocument(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user := &User{}
		user.setName("RavenDB")

		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	patchRequest := &ravendb.PatchRequest{
		Script: `this.name = "Patched"`,
	}
	patchOperation, err := ravendb.NewPatchOperation("users/1", nil, patchRequest, nil, false)
	assert.NoError(t, err)
	patchResult, err := store.Operations().SendPatchOperation(patchOperation, nil)
	assert.NoError(t, err)
	assert.Equal(t, patchResult.Status, ravendb.PatchStatusPatched)

	{
		session := openSessionMust(t, store)
		var loadedUser *User
		err = session.Load(&loadedUser, "users/1")
		assert.NoError(t, err)
		assert.Equal(t, *loadedUser.Name, "Patched")
		session.Close()
	}
}

func patchTestCanWaitForIndexAfterPatch(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	usersByName := NewUsers_ByName()
	err = usersByName.Execute(store, nil, "")
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		user := &User{}
		user.setName("RavenDB")

		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		builder := func(x *ravendb.IndexesWaitOptsBuilder) {
			x.WaitForIndexes("Users/ByName")
		}
		session.Advanced().WaitForIndexesAfterSaveChanges(builder)

		var user *User
		err = session.Load(&user, "users/1")
		assert.NoError(t, err)

		err = session.Advanced().Patch(user, "name", "New Name")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}
}

func patchTestcanPatchManyDocuments(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user := &User{}
		user.setName("RavenDB")

		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		// TODO: we crash if it's Query() and not QueryType()
		// (need to validate and not crash)
		clazz := reflect.TypeOf(&User{})
		q := session.QueryCollectionForType(clazz)
		lazy, err := q.CountLazily()
		assert.NoError(t, err)
		var n int
		err = lazy.GetValue(&n)
		assert.NoError(t, err)
		assert.Equal(t, n, 1)

		session.Close()
	}

	operation := ravendb.NewPatchByQueryOperation("from Users update {  this.name= \"Patched\"  }")
	op, err := store.Operations().SendAsync(operation, nil)
	assert.NoError(t, err)
	err = op.WaitForCompletion()
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		var loadedUser *User
		err = session.Load(&loadedUser, "users/1")
		assert.NoError(t, err)
		assert.Equal(t, *loadedUser.Name, "Patched")
		session.Close()
	}
}

func patchTestthrowsOnInvalidScript(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user := &User{}
		user.setName("RavenDB")

		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	operation := ravendb.NewPatchByQueryOperation("from Users update {  throw 5 }")

	op, err := store.Operations().SendAsync(operation, nil)
	assert.NoError(t, err)

	err = op.WaitForCompletion()
	assert.Error(t, err)
	// TODO: make sure it's an instance of JavaScriptException ? Currently is RavenError
	assert.True(t, strings.Contains(err.Error(), "Raven.Client.Exceptions.Documents.Patching.JavaScriptException"))

}

func TestPatch(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// order matches Java tests
	patchTestcanPatchManyDocuments(t, driver)
	patchTestthrowsOnInvalidScript(t, driver)
	patchTestcanPatchSingleDocument(t, driver)

	// TODO: not in order of Java
	patchTestCanWaitForIndexAfterPatch(t, driver)
}
