package ravendb

import (
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func deleteDocumentCommandTest_canDeleteDocument(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user := NewUser()
		user.setName("Marcin")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}
	command := NewDeleteDocumentCommand("users/1", nil)
	err = store.getRequestExecutor().executeCommand(command)
	assert.NoError(t, err)
	{
		session := openSessionMust(t, store)
		loadedUserI, err := session.load(getTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		loadedUser := loadedUserI.(*User)
		assert.Nil(t, loadedUser)
		session.Close()
	}
}

func deleteDocumentCommandTest_canDeleteDocumentByEtag(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	var changeVector *string
	{
		session := openSessionMust(t, store)
		user := NewUser()
		user.setName("Marcin")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		changeVector, err = session.advanced().getChangeVectorFor(user)
		assert.NoError(t, err)
		session.Close()
	}

	{
		session := openSessionMust(t, store)
		loadedUserI, err := session.load(getTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		loadedUser := loadedUserI.(*User)
		assert.NotNil(t, loadedUser)
		loadedUser.setAge(5)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	command := NewDeleteDocumentCommand("users/1", changeVector)
	err = store.getRequestExecutor().executeCommand(command)
	assert.Error(t, err)
	_ = err.(*ConcurrencyException)
}

func TestDeleteDocumentCommand(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer func() {
		r := recover()
		destroyDriver()
		if r != nil {
			fmt.Printf("Panic: '%v'\n", r)
			debug.PrintStack()
			t.Fail()
		}
	}()

	// follows execution order of java tests
	deleteDocumentCommandTest_canDeleteDocument(t)
	deleteDocumentCommandTest_canDeleteDocumentByEtag(t)
}
