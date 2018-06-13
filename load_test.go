package ravendb

import (
	"strconv"
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

func loadTest_loadCanUseCache(t *testing.T) {
	store := getDocumentStoreMust(t)
	{
		session := openSessionMust(t, store)
		user := NewUser()
		user.setName("RavenDB")

		err := session.StoreEntityWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
	}

	{
		newSession := openSessionMust(t, store)
		result, err := newSession.load(getTypeOfValue(NewUser()), "users/1")
		assert.NoError(t, err)
		user := result.(*User)
		assert.NotNil(t, user)
	}

	{
		newSession := openSessionMust(t, store)
		result, err := newSession.load(getTypeOfValue(NewUser()), "users/1")
		assert.NoError(t, err)
		user := result.(*User)
		assert.NotNil(t, user)
	}
}

func loadTest_loadDocumentById(t *testing.T) {
	store := getDocumentStoreMust(t)
	{
		session := openSessionMust(t, store)
		user := NewUser()
		user.setName("RavenDB")

		err := session.StoreEntityWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
	}

	{
		newSession := openSessionMust(t, store)
		result, err := newSession.load(getTypeOfValue(NewUser()), "users/1")
		assert.NoError(t, err)
		user := result.(*User)
		assert.NotNil(t, user)
		assert.Equal(t, "RavenDB", *user.getName())
	}
}

func loadTest_loadDocumentsByIds(t *testing.T) {
	store := getDocumentStoreMust(t)
	{
		session := openSessionMust(t, store)
		user1 := NewUser()
		user1.setName("RavenDB")

		user2 := NewUser()
		user2.setName("Hibernating Rhinos")

		err := session.StoreEntityWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.StoreEntityWithID(user2, "users/2")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
	}

	{
		newSession := openSessionMust(t, store)
		users, err := newSession.loadMulti(getTypeOfValue(NewUser()), []string{"users/1", "users/2"})
		assert.NoError(t, err)
		assert.Equal(t, 2, len(users))
	}
}

func loadTest_loadNullShouldReturnNull(t *testing.T) {
	store := getDocumentStoreMust(t)
	{
		session := openSessionMust(t, store)
		user1 := NewUser()
		user1.setName("Tony Montana")

		user2 := NewUser()
		user2.setName("Tony Soprano")

		err := session.StoreEntity(user1)
		assert.NoError(t, err)
		err = session.StoreEntity(user2)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)
	}

	{
		newSession := openSessionMust(t, store)
		user1, err := newSession.load(getTypeOfValue(&User{}), "")
		assert.NoError(t, err)
		assert.Nil(t, user1)
	}
}

func loadTest_loadMultiIdsWithNullShouldReturnDictionaryWithoutNulls(t *testing.T) {
}

func loadTest_loadDocumentWithINtArrayAndLongArray(t *testing.T) {
}

func loadTest_shouldLoadManyIdsAsPostRequest(t *testing.T) {

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
			err := session.StoreEntityWithID(user, id)
			assert.NoError(t, err)
		}

		err := session.SaveChanges()
		assert.NoError(t, err)
	}

	{
		session := openSessionMust(t, store)
		users, err := session.loadMulti(getTypeOfValue(&User{}), ids)
		assert.NoError(t, err)
		assert.NotNil(t, users)
		result := users["users/77"]
		user := result.(*User)
		assert.NotNil(t, user)
		assert.Equal(t, "users/77", user.ID)
	}
}

func loadTest_loadStartsWith(t *testing.T) {
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
	if true {
		loadTest_loadDocumentById(t)
		loadTest_loadNullShouldReturnNull(t)
	}

	//TODO: fails for now
	//loadTest_loadDocumentsByIds(t)

	//TODO: fails for now
	//loadTest_shouldLoadManyIdsAsPostRequest(t)

	if true {
		loadTest_loadStartsWith(t)
		loadTest_loadMultiIdsWithNullShouldReturnDictionaryWithoutNulls(t)
		loadTest_loadDocumentWithINtArrayAndLongArray(t)
		loadTest_loadCanUseCache(t)
	}
}