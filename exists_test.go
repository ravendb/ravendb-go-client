package ravendb

import (
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/assert"
)

func existsTest_checkIfDocumentExists(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		assert.NoError(t, err)
		idan := NewUser()
		idan.setName("Idan")

		shalom := NewUser()
		shalom.setName("Shalom")

		err = session.StoreEntityWithID(idan, "users/1")
		assert.NoError(t, err)
		err = session.StoreEntityWithID(shalom, "users/2")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		session := openSessionMust(t, store)
		assert.NoError(t, err)
		ok, err := session.advanced().exists("users/1")
		assert.NoError(t, err)
		assert.True(t, ok)

		ok, err = session.advanced().exists("users/10")
		assert.NoError(t, err)
		assert.False(t, ok)

		_, err = session.load(getTypeOf(NewUser()), "users/2")
		assert.NoError(t, err)
		ok, err = session.advanced().exists("users/2")
		assert.NoError(t, err)
		assert.True(t, ok)
		session.Close()
	}
}

func TestExists(t *testing.T) {
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

	existsTest_checkIfDocumentExists(t)
}
