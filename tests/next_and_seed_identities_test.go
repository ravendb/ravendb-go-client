package tests

import (
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func nextAndSeedIdentitiesTest_nextIdentityFor(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user := &User{}
		user.setLastName("Adi")

		err = session.StoreWithID(user, "users|")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	command := ravendb.NewNextIdentityForCommand("users")
	err = store.GetRequestExecutor().ExecuteCommand(command)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		user := &User{}
		user.setLastName("Avivi")

		err = session.StoreWithID(user, "users|")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		var entityWithId1, entityWithId2, entityWithId3, entityWithId4 *User
		session := openSessionMust(t, store)
		err = session.Load(&entityWithId1, "users/1")
		assert.NoError(t, err)
		err = session.Load(&entityWithId2, "users/2")
		assert.NoError(t, err)
		err = session.Load(&entityWithId3, "users/3")
		assert.NoError(t, err)
		err = session.Load(&entityWithId4, "users/4")
		assert.NoError(t, err)

		assert.NotNil(t, entityWithId1)
		assert.NotNil(t, entityWithId3)
		assert.Nil(t, entityWithId2)
		assert.Nil(t, entityWithId4)

		assert.Equal(t, *entityWithId1.LastName, "Adi")
		assert.Equal(t, *entityWithId3.LastName, "Avivi")
		session.Close()
	}
}

func nextAndSeedIdentitiesTest_seedIdentityFor(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user := &User{}
		user.setLastName("Adi")

		err = session.StoreWithID(user, "users|")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	command := ravendb.NewSeedIdentityForCommand("users", 1990)
	err = store.GetRequestExecutor().ExecuteCommand(command)
	assert.NoError(t, err)
	result := command.Result
	assert.Equal(t, result, 1990)

	{
		session := openSessionMust(t, store)
		user := &User{}
		user.setLastName("Avivi")

		err = session.StoreWithID(user, "users|")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		session := openSessionMust(t, store)
		var entityWithId1, entityWithId2, entityWithId1990, entityWithId1991, entityWithId1992 *User
		err = session.Load(&entityWithId1, "users/1")
		assert.NoError(t, err)

		err = session.Load(&entityWithId2, "users/2")
		assert.NoError(t, err)

		err = session.Load(&entityWithId1990, "users/1990")
		assert.NoError(t, err)

		err = session.Load(&entityWithId1991, "users/1991")
		assert.NoError(t, err)

		err = session.Load(&entityWithId1992, "users/1992")
		assert.NoError(t, err)

		assert.NotNil(t, entityWithId1)
		assert.NotNil(t, entityWithId1991)

		assert.Nil(t, entityWithId2)
		assert.Nil(t, entityWithId1990)
		assert.Nil(t, entityWithId1992)

		assert.Equal(t, *entityWithId1.LastName, "Adi")
		assert.Equal(t, *entityWithId1991.LastName, "Avivi")
		session.Close()
	}

	command = ravendb.NewSeedIdentityForCommand("users", 1975)
	err = store.GetRequestExecutor().ExecuteCommand(command)
	assert.NoError(t, err)
	assert.Equal(t, command.Result, 1991)

	{
		op := ravendb.NewGetIdentitiesOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)

		identites := op.Command.Result
		n := identites["users|"]
		assert.Equal(t, n, 1991)
	}
}

func TestNextAndSeedIdentities(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	nextAndSeedIdentitiesTest_nextIdentityFor(t, driver)
	nextAndSeedIdentitiesTest_seedIdentityFor(t, driver)
}
