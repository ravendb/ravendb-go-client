package ravendb

import (
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
		result := newSession.load(getTypeOfValue(NewUser()), "users/1")
		user := result.(*User)
		assert.NotNil(t, user)
	}

	{
		newSession := openSessionMust(t, store)
		result := newSession.load(getTypeOfValue(NewUser()), "users/1")
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
		result := newSession.load(getTypeOfValue(NewUser()), "users/1")
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
		users := newSession.loadMulti(getTypeOfValue(NewUser()), "users/1", "users/2")
		assert.Equal(t, 2, len(users))
	}
}

func loadTest_loadNullShouldReturnNull(t *testing.T) {
}
func loadTest_loadMultiIdsWithNullShouldReturnDictionaryWithoutNulls(t *testing.T) {
}
func loadTest_loadDocumentWithINtArrayAndLongArray(t *testing.T) {
}
func loadTest_shouldLoadManyIdsAsPostRequest(t *testing.T) {
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
	loadTest_loadDocumentById(t)
	loadTest_loadNullShouldReturnNull(t)
	//TODO: failing for now
	//loadTest_loadDocumentsByIds(t)
	loadTest_shouldLoadManyIdsAsPostRequest(t)
	loadTest_loadStartsWith(t)
	loadTest_loadMultiIdsWithNullShouldReturnDictionaryWithoutNulls(t)
	loadTest_loadDocumentWithINtArrayAndLongArray(t)
	loadTest_loadCanUseCache(t)
}
