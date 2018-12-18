package tests

import (
	"strings"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func patchTestcanPatchSingleDocument(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
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

	patchOperation := ravendb.NewPatchOperation("users/1", nil,
		ravendb.PatchRequest_forScript("this.name = \"Patched\""), nil, false)
	err = store.Operations().Send(patchOperation)
	assert.NoError(t, err)
	status := patchOperation.Command.Result
	assert.Equal(t, status.GetStatus(), ravendb.PatchStatus_PATCHED)

	{
		session := openSessionMust(t, store)
		var loadedUser *User
		err = session.Load(&loadedUser, "users/1")
		assert.NoError(t, err)
		assert.Equal(t, *loadedUser.Name, "Patched")
		session.Close()
	}
}

func patchTestcanPatchManyDocuments(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
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

	operation := ravendb.NewPatchByQueryOperation("from Users update {  this.name= \"Patched\"  }")
	op, err := store.Operations().SendAsync(operation)
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
	store := getDocumentStoreMust(t, driver)
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

	op, err := store.Operations().SendAsync(operation)
	assert.NoError(t, err)

	err = op.WaitForCompletion()
	assert.Error(t, err)
	// TODO: make sure it's an instance of JavaScriptException ? Currently is RavenError
	assert.True(t, strings.Contains(err.Error(), "Raven.Client.Exceptions.Documents.Patching.JavaScriptException"))

}

func TestPatch(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// order matches Java tests
	patchTestcanPatchManyDocuments(t, driver)
	patchTestthrowsOnInvalidScript(t, driver)
	patchTestcanPatchSingleDocument(t, driver)
}
