package ravendb

import (
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

func patchTestcanPatchSingleDocument(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
	{
		session := openSessionMust(t, store)
		user := NewUser()
		user.setName("RavenDB")

		err = session.StoreEntityWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
	}
}

func patchTestthrowsOnInvalidScript(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	{
		session := openSessionMust(t, store)
		user := NewUser()
		user.setName("RavenDB")

		err = session.StoreEntityWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
	}
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

	// TODO: needs PatchByQueryOperation
	//patchTestthrowsOnInvalidScript(t)
	//patchTestcanPatchSingleDocument(t)
}
