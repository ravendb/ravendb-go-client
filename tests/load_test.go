package tests

import (
	"strconv"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func loadTestLoadCanUseCache(t *testing.T, driver *RavenTestDriver) {
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

	{
		newSession := openSessionMust(t, store)
		var user *User
		err = newSession.Load(&user, "users/1")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		newSession.Close()
	}

	{
		newSession := openSessionMust(t, store)
		var user *User
		err = newSession.Load(&user, "users/1")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		newSession.Close()
	}
}

func loadTestLoadDocumentById(t *testing.T, driver *RavenTestDriver) {
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

	{
		newSession := openSessionMust(t, store)
		var user *User
		err = newSession.Load(&user, "users/1")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, "RavenDB", *user.Name)
		newSession.Close()
	}
}

func loadTestLoaddocumentsByIDs(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user1 := &User{}
		user1.setName("RavenDB")

		user2 := &User{}
		user2.setName("Hibernating Rhinos")

		err = session.StoreWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.StoreWithID(user2, "users/2")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		newSession := openSessionMust(t, store)
		users := map[string]*User{}
		err = newSession.LoadMulti(users, []string{"users/1", "users/2"})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(users))
		newSession.Close()
	}
}

func loadTestLoadNullShouldReturnNull(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user1 := &User{}
		user1.setName("Tony Montana")

		user2 := &User{}
		user2.setName("Tony Soprano")

		err = session.Store(user1)
		assert.NoError(t, err)
		err = session.Store(user2)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		newSession := openSessionMust(t, store)
		var user1 *User
		err = newSession.Load(&user1, "")
		assert.NoError(t, err)
		assert.Nil(t, user1)
		newSession.Close()
	}
}

func loadTestLoadMultiIdsWithNullShouldReturnDictionaryWithoutNulls(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user1 := &User{}
		user1.setName("Tony Montana")

		user2 := &User{}
		user2.setName("Tony Soprano")

		err = session.StoreWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.StoreWithID(user2, "users/2")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}
	{
		newSession := openSessionMust(t, store)
		orderedArrayOfIdsWithNull := []string{"users/1", "", "users/2", ""}
		users1 := map[string]*User{}
		err = newSession.LoadMulti(users1, orderedArrayOfIdsWithNull)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(users1))

		user1 := users1["users/1"]
		assert.NotNil(t, user1)

		user2 := users1["users/2"]
		assert.NotNil(t, user2)
		newSession.Close()
	}
}

func loadTestLoadDocumentWithIntArrayAndLongArray(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		geek1 := &GeekPerson{
			Name:                    "Bebop",
			FavoritePrimes:          []int{13, 43, 443, 997},
			FavoriteVeryLargePrimes: []int64{5000000029, 5000000039},
		}

		err = session.StoreWithID(geek1, "geeks/1")
		assert.NoError(t, err)

		geek2 := &GeekPerson{
			Name:                    "Rocksteady",
			FavoritePrimes:          []int{2, 3, 5, 7},
			FavoriteVeryLargePrimes: []int64{999999999989},
		}

		err = session.StoreWithID(geek2, "geeks/2")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		newSession := openSessionMust(t, store)
		var geek1 *GeekPerson
		err = newSession.Load(&geek1, "geeks/1")
		assert.NoError(t, err)

		var geek2 *GeekPerson
		err = newSession.Load(&geek2, "geeks/2")
		assert.NoError(t, err)

		assert.Equal(t, 43, geek1.FavoritePrimes[1])
		assert.Equal(t, int64(5000000039), geek1.FavoriteVeryLargePrimes[1])

		assert.Equal(t, 7, geek2.FavoritePrimes[3])
		assert.Equal(t, int64(999999999989), geek2.FavoriteVeryLargePrimes[0])
		newSession.Close()
	}
}

func loadTestShouldLoadManyIdsAsPostRequest(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	var ids []string

	{
		session := openSessionMust(t, store)
		// Length of all the ids together should be larger than 1024 for POST request
		for i := 0; i < 200; i++ {
			id := "users/" + strconv.Itoa(i)
			ids = append(ids, id)

			user := &User{}
			user.setName("Person " + strconv.Itoa(i))
			err = session.StoreWithID(user, id)
			assert.NoError(t, err)
		}

		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		session := openSessionMust(t, store)
		users := map[string]*User{}
		err = session.LoadMulti(users, ids)
		assert.NoError(t, err)
		assert.NotNil(t, users)
		user := users["users/77"]
		assert.NotNil(t, user)
		name := *user.Name
		assert.Equal(t, "Person 77", name)
		assert.Equal(t, "users/77", user.ID)
		session.Close()
	}
}

func loadTestLoadStartsWith(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		createUser := func(id string) *User {
			u := &User{}
			u.ID = id
			err = session.Store(u)
			assert.NoError(t, err)
			return u
		}

		createUser("Aaa")
		createUser("Abc")
		createUser("Afa")
		createUser("Ala")
		createUser("Baa")

		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		newSession := openSessionMust(t, store)
		var users []*User
		args := &ravendb.StartsWithArgs{
			StartsWith: "A",
		}
		err = newSession.Advanced().LoadStartingWith(&users, args)
		assert.NoError(t, err)

		userIDs := []string{"Aaa", "Abc", "Afa", "Ala"}
		for _, user := range users {
			assert.True(t, stringArrayContains(userIDs, user.ID))
		}

		users = nil
		args = &ravendb.StartsWithArgs{
			StartsWith: "A",
			Start:      1,
			PageSize:   2,
		}
		err = newSession.Advanced().LoadStartingWith(&users, args)

		userIDs = []string{"Abc", "Afa"}
		for _, user := range users {
			assert.True(t, stringArrayContains(userIDs, user.ID))
		}
		newSession.Close()
	}
}

func TestLoad(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	loadTestLoadDocumentById(t, driver)
	loadTestLoadNullShouldReturnNull(t, driver)
	loadTestLoaddocumentsByIDs(t, driver)
	loadTestShouldLoadManyIdsAsPostRequest(t, driver)
	loadTestLoadStartsWith(t, driver)
	loadTestLoadMultiIdsWithNullShouldReturnDictionaryWithoutNulls(t, driver)
	loadTestLoadDocumentWithIntArrayAndLongArray(t, driver)
	loadTestLoadCanUseCache(t, driver)
}
