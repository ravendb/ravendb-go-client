package tests

import (
	"strconv"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func loadTest_loadCanUseCache(t *testing.T, driver *RavenTestDriver) {
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

func loadTest_loadDocumentById(t *testing.T, driver *RavenTestDriver) {
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

func loadTest_loaddocumentsByIDs(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
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

func loadTest_loadNullShouldReturnNull(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
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

func loadTest_loadMultiIdsWithNullShouldReturnDictionaryWithoutNulls(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
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

func loadTest_loadDocumentWithIntArrayAndLongArray(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		geek1 := NewGeekPerson()
		geek1.setName("Bebop")
		geek1.setFavoritePrimes([]int{13, 43, 443, 997})
		geek1.setFavoriteVeryLargePrimes([]int64{5000000029, 5000000039})

		err = session.StoreWithID(geek1, "geeks/1")
		assert.NoError(t, err)

		geek2 := NewGeekPerson()
		geek2.setName("Rocksteady")
		geek2.setFavoritePrimes([]int{2, 3, 5, 7})
		geek2.setFavoriteVeryLargePrimes([]int64{999999999989})

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

		assert.Equal(t, 43, geek1.getFavoritePrimes()[1])
		assert.Equal(t, int64(5000000039), geek1.getFavoriteVeryLargePrimes()[1])

		assert.Equal(t, 7, geek2.getFavoritePrimes()[3])
		assert.Equal(t, int64(999999999989), geek2.getFavoriteVeryLargePrimes()[0])
		newSession.Close()
	}
}

func loadTest_shouldLoadManyIdsAsPostRequest(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
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

func loadTest_loadStartsWith(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
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
			assert.True(t, ravendb.StringArrayContains(userIDs, user.ID))
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
			assert.True(t, ravendb.StringArrayContains(userIDs, user.ID))
		}
		newSession.Close()
	}
}

func TestLoad(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	loadTest_loadDocumentById(t, driver)
	loadTest_loadNullShouldReturnNull(t, driver)
	loadTest_loaddocumentsByIDs(t, driver)
	loadTest_shouldLoadManyIdsAsPostRequest(t, driver)
	loadTest_loadStartsWith(t, driver)
	loadTest_loadMultiIdsWithNullShouldReturnDictionaryWithoutNulls(t, driver)
	loadTest_loadDocumentWithIntArrayAndLongArray(t, driver)
	loadTest_loadCanUseCache(t, driver)
}
