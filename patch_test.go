package ravendb

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func patchTestcanPatchSingleDocument(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user := NewUser()
		user.setName("RavenDB")

		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	patchOperation := NewPatchOperation("users/1", nil,
		PatchRequest_forScript("this.name = \"Patched\""), nil, false)
	err = store.Operations().Send(patchOperation)
	assert.NoError(t, err)
	status := patchOperation.Command.Result
	assert.Equal(t, status.getStatus(), PatchStatus_PATCHED)

	{
		session := openSessionMust(t, store)
		loadedUserI, err := session.Load(GetTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		loadedUser := loadedUserI.(*User)
		assert.Equal(t, *loadedUser.getName(), "Patched")
		session.Close()
	}
}

func patchTestcanPatchManyDocuments(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user := NewUser()
		user.setName("RavenDB")

		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	operation := NewPatchByQueryOperation("from Users update {  this.name= \"Patched\"  }")
	op, err := store.Operations().sendAsync(operation)
	assert.NoError(t, err)
	err = op.waitForCompletion()
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		loadedUserI, err := session.Load(GetTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		loadedUser := loadedUserI.(*User)
		assert.Equal(t, *loadedUser.getName(), "Patched")
		session.Close()
	}
}

func patchTestthrowsOnInvalidScript(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user := NewUser()
		user.setName("RavenDB")

		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	operation := NewPatchByQueryOperation("from Users update {  throw 5 }")

	op, err := store.Operations().sendAsync(operation)
	assert.NoError(t, err)

	err = op.waitForCompletion()
	assert.Error(t, err)
	// TODO: make sure it's an instance of JavaScriptException ? Currently is RavenException
	assert.True(t, strings.Contains(err.Error(), "Raven.Client.Exceptions.Documents.Patching.JavaScriptException"))

}

func TestPatch(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// order matches Java tests
	patchTestcanPatchManyDocuments(t)
	patchTestthrowsOnInvalidScript(t)
	patchTestcanPatchSingleDocument(t)
}
