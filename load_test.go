package ravendb

import (
	"strconv"
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

func loadTest_loadCanUseCache(t *testing.T) {
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

	{
		newSession := openSessionMust(t, store)
		result, err := newSession.load(getTypeOf(NewUser()), "users/1")
		assert.NoError(t, err)
		user := result.(*User)
		assert.NotNil(t, user)
	}

	{
		newSession := openSessionMust(t, store)
		result, err := newSession.load(getTypeOf(NewUser()), "users/1")
		assert.NoError(t, err)
		user := result.(*User)
		assert.NotNil(t, user)
	}
}

func loadTest_loadDocumentById(t *testing.T) {
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

	{
		newSession := openSessionMust(t, store)
		result, err := newSession.load(getTypeOf(NewUser()), "users/1")
		assert.NoError(t, err)
		user := result.(*User)
		assert.NotNil(t, user)
		assert.Equal(t, "RavenDB", *user.getName())
	}
}

func loadTest_loadDocumentsByIds(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	{
		session := openSessionMust(t, store)
		user1 := NewUser()
		user1.setName("RavenDB")

		user2 := NewUser()
		user2.setName("Hibernating Rhinos")

		err = session.StoreEntityWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.StoreEntityWithID(user2, "users/2")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
	}

	{
		newSession := openSessionMust(t, store)
		users, err := newSession.loadMulti(getTypeOf(NewUser()), []string{"users/1", "users/2"})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(users))
	}
}

func loadTest_loadNullShouldReturnNull(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	{
		session := openSessionMust(t, store)
		user1 := NewUser()
		user1.setName("Tony Montana")

		user2 := NewUser()
		user2.setName("Tony Soprano")

		err = session.StoreEntity(user1)
		assert.NoError(t, err)
		err = session.StoreEntity(user2)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)
	}

	{
		newSession := openSessionMust(t, store)
		user1, err := newSession.load(getTypeOf(&User{}), "")
		assert.NoError(t, err)
		assert.Nil(t, user1)
	}
}

func loadTest_loadMultiIdsWithNullShouldReturnDictionaryWithoutNulls(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	{
		session := openSessionMust(t, store)
		user1 := NewUser()
		user1.setName("Tony Montana")

		user2 := NewUser()
		user2.setName("Tony Soprano")

		err = session.StoreEntityWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.StoreEntityWithID(user2, "users/2")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
	}
	{
		newSession := openSessionMust(t, store)
		orderedArrayOfIdsWithNull := []string{"users/1", "", "users/2", ""}
		users1, err := newSession.loadMulti(getTypeOf(&User{}), orderedArrayOfIdsWithNull)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(users1))

		ruser1 := users1["users/1"]
		user1 := ruser1.(*User)
		assert.NotNil(t, user1)

		ruser2 := users1["users/2"]
		user2 := ruser2.(*User)
		assert.NotNil(t, user2)
	}
}

func loadTest_loadDocumentWithINtArrayAndLongArray(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)

	{
		session := openSessionMust(t, store)
		geek1 := NewGeekPerson()
		geek1.setName("Bebop")
		geek1.setFavoritePrimes([]int{13, 43, 443, 997})
		geek1.setFavoriteVeryLargePrimes([]int64{5000000029, 5000000039})

		err = session.StoreEntityWithID(geek1, "geeks/1")
		assert.NoError(t, err)

		geek2 := NewGeekPerson()
		geek2.setName("Rocksteady")
		geek2.setFavoritePrimes([]int{2, 3, 5, 7})
		geek2.setFavoriteVeryLargePrimes([]int64{999999999989})

		err = session.StoreEntityWithID(geek2, "geeks/2")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
	}

	{
		newSession := openSessionMust(t, store)
		geek1i, err := newSession.load(getTypeOf(&GeekPerson{}), "geeks/1")
		assert.NoError(t, err)
		geek1 := geek1i.(*GeekPerson)

		geek2i, err := newSession.load(getTypeOf(&GeekPerson{}), "geeks/2")
		assert.NoError(t, err)
		geek2 := geek2i.(*GeekPerson)

		assert.Equal(t, 43, geek1.getFavoritePrimes()[1])
		assert.Equal(t, int64(5000000039), geek1.getFavoriteVeryLargePrimes()[1])

		assert.Equal(t, 7, geek2.getFavoritePrimes()[3])
		assert.Equal(t, int64(999999999989), geek2.getFavoriteVeryLargePrimes()[0])

	}
}

func loadTest_shouldLoadManyIdsAsPostRequest(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	var ids []string

	{
		session := openSessionMust(t, store)
		// Length of all the ids together should be larger than 1024 for POST request
		for i := 0; i < 200; i++ {
			id := "users/" + strconv.Itoa(i)
			ids = append(ids, id)

			user := NewUser()
			user.setName("Person " + strconv.Itoa(i))
			err = session.StoreEntityWithID(user, id)
			assert.NoError(t, err)
		}

		err = session.SaveChanges()
		assert.NoError(t, err)
	}

	{
		session := openSessionMust(t, store)
		users, err := session.loadMulti(getTypeOf(&User{}), ids)
		assert.NoError(t, err)
		assert.NotNil(t, users)
		result := users["users/77"]
		user := result.(*User)
		assert.NotNil(t, user)
		name := *user.Name
		assert.Equal(t, "Person 77", name)
		assert.Equal(t, "users/77", user.ID)
	}
}

func loadTest_loadStartsWith(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)

	{
		session := openSessionMust(t, store)
		createUser := func(id string) *User {
			u := NewUser()
			u.setId(id)
			err = session.StoreEntity(u)
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
	}

	{
		newSession := openSessionMust(t, store)
		usersi, err := newSession.advanced().loadStartingWith(getTypeOf(&User{}), "A")
		assert.NoError(t, err)

		userIDs := []string{"Aaa", "Abc", "Afa", "Ala"}
		for _, useri := range usersi {
			user := useri.(*User)
			assert.True(t, stringArrayContains(userIDs, user.ID))
		}

		usersi, err = newSession.advanced().loadStartingWithFull(getTypeOf(&User{}), "A", "", 1, 2, "", "")

		userIDs = []string{"Abc", "Afa"}
		for _, useri := range usersi {
			user := useri.(*User)
			assert.True(t, stringArrayContains(userIDs, user.ID))
		}
	}
}

func TestLoad(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	if useProxy() {
		proxy.ChangeLogFile("trace_load_go.txt")
	}

	createTestDriver()
	defer deleteTestDriver()

	// matches order of Java tests
	loadTest_loadDocumentById(t)
	loadTest_loadNullShouldReturnNull(t)
	loadTest_loadDocumentsByIds(t)
	loadTest_shouldLoadManyIdsAsPostRequest(t)
	loadTest_loadStartsWith(t)
	loadTest_loadMultiIdsWithNullShouldReturnDictionaryWithoutNulls(t)
	loadTest_loadDocumentWithINtArrayAndLongArray(t)
	loadTest_loadCanUseCache(t)
}
