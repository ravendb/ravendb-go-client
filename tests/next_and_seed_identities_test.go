package tests

import (
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func nextAndSeedIdentitiesTest_nextIdentityFor(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
		session := openSessionMust(t, store)
		entityWithId1I, err := session.LoadOld(ravendb.GetTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		entityWithId1 := entityWithId1I.(*User)
		entityWithId2I, err := session.LoadOld(ravendb.GetTypeOf(&User{}), "users/2")
		assert.NoError(t, err)
		entityWithId2 := entityWithId2I.(*User)
		entityWithId3I, err := session.LoadOld(ravendb.GetTypeOf(&User{}), "users/3")
		assert.NoError(t, err)
		entityWithId3 := entityWithId3I.(*User)
		entityWithId4I, err := session.LoadOld(ravendb.GetTypeOf(&User{}), "users/4")
		assert.NoError(t, err)
		entityWithId4 := entityWithId4I.(*User)

		assert.NotNil(t, entityWithId1)
		assert.NotNil(t, entityWithId3)
		assert.Nil(t, entityWithId2)
		assert.Nil(t, entityWithId4)

		assert.Equal(t, *entityWithId1.LastName, "Adi")
		assert.Equal(t, *entityWithId3.LastName, "Avivi")
		session.Close()
	}
}

func nextAndSeedIdentitiesTest_seedIdentityFor(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
		entityWithId1I, err := session.LoadOld(ravendb.GetTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		entityWithId1 := entityWithId1I.(*User)

		entityWithId2I, err := session.LoadOld(ravendb.GetTypeOf(&User{}), "users/2")
		assert.NoError(t, err)
		entityWithId2 := entityWithId2I.(*User)

		entityWithId1990I, err := session.LoadOld(ravendb.GetTypeOf(&User{}), "users/1990")
		assert.NoError(t, err)
		entityWithId1990 := entityWithId1990I.(*User)

		entityWithId1991I, err := session.LoadOld(ravendb.GetTypeOf(&User{}), "users/1991")
		assert.NoError(t, err)
		entityWithId1991 := entityWithId1991I.(*User)

		entityWithId1992I, err := session.LoadOld(ravendb.GetTypeOf(&User{}), "users/1992")
		assert.NoError(t, err)
		entityWithId1992 := entityWithId1992I.(*User)

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

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	nextAndSeedIdentitiesTest_nextIdentityFor(t)
	nextAndSeedIdentitiesTest_seedIdentityFor(t)
}
