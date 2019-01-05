package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func go1Test(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	session := openSessionMust(t, store)
	user := User{}

	// check validation of arguments to Store and Delete

	// can't store/delete nil
	err = session.Store(nil)
	assert.Error(t, err)
	err = session.StoreWithID(nil, "users/1")
	assert.Error(t, err)
	err = session.DeleteEntity(nil)
	assert.Error(t, err)

	// can't store/delete struct
	err = session.Store(user)
	assert.Error(t, err)
	err = session.StoreWithID(user, "users/1")
	assert.Error(t, err)
	err = session.DeleteEntity(user)
	assert.Error(t, err)

	// can't store/delete **struct (double pointer values)
	ptrUser := &user
	err = session.Store(&ptrUser)
	assert.Error(t, err)
	err = session.StoreWithID(&ptrUser, "users/1")
	assert.Error(t, err)
	err = session.DeleteEntity(&ptrUser)
	assert.Error(t, err)

	// can't store/delete nil pointer
	var user2 *User
	err = session.Store(user2)
	assert.Error(t, err)
	err = session.StoreWithID(user2, "users/1")
	assert.Error(t, err)
	err = session.DeleteEntity(user2)
	assert.Error(t, err)
}

func TestGo1(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	go1Test(t, driver)
}
