package ravendb

import (
	"strings"
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
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

		err = session.StoreEntityWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
	}

	patchOperation := NewPatchOperation("users/1", nil,
		PatchRequest_forScript("this.name = \"Patched\""), nil, false)
	err = store.operations().send(patchOperation)
	assert.NoError(t, err)
	status := patchOperation.Command.Result
	assert.Equal(t, status.getStatus(), PatchStatus_PATCHED)

	{
		session := openSessionMust(t, store)
		loadedUserI, err := session.load(getTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		loadedUser := loadedUserI.(*User)
		assert.Equal(t, *loadedUser.getName(), "Patched")
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

		err = session.StoreEntityWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
	}

	operation := NewPatchByQueryOperation("from Users update {  this.name= \"Patched\"  }")
	op, err := store.operations().sendAsync(operation)
	assert.NoError(t, err)
	err = op.waitForCompletion()
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		loadedUserI, err := session.load(getTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		loadedUser := loadedUserI.(*User)
		assert.Equal(t, *loadedUser.getName(), "Patched")
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

		err = session.StoreEntityWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
	}

	operation := NewPatchByQueryOperation("from Users update {  throw 5 }")

	op, err := store.operations().sendAsync(operation)
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
	if useProxy() {
		proxy.ChangeLogFile("trace_patch_go.txt")
	}

	createTestDriver()
	defer deleteTestDriver()

	// order matches Java tests
	patchTestcanPatchManyDocuments(t)
	patchTestthrowsOnInvalidScript(t)
	patchTestcanPatchSingleDocument(t)
}
