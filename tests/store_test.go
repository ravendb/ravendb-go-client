package tests

import (
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func storeTestRefreshTest(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user := &User{}
		user.setName("RavenDB")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		{
			innerSession := openSessionMust(t, store)
			var innerUser *User
			err = innerSession.Load(&innerUser, "users/1")
			assert.NoError(t, err)

			innerUser.setName("RavenDB 4.0")
			err = innerSession.SaveChanges()
			assert.NoError(t, err)
		}

		session.Advanced().Refresh(user)

		name := *user.Name
		assert.Equal(t, name, "RavenDB 4.0")
		session.Close()
	}
}

func storeTestStoreDocument(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user := &User{}
		user.setName("RavenDB")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		user = nil
		err = session.Load(&user, "users/1")
		assert.NoError(t, err)
		assert.NotNil(t, user)
		name := *user.Name
		assert.Equal(t, name, "RavenDB")
		session.Close()
	}
}

func storeTestStoreDocuments(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user1 := &User{}
		user1.setName("RavenDB")
		err = session.StoreWithID(user1, "users/1")
		assert.NoError(t, err)

		user2 := &User{}
		user2.setName("Hibernating Rhinos")
		err = session.StoreWithID(user2, "users/2")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		users := map[string]*User{}
		err = session.LoadMulti(users, []string{"users/1", "users/2"})
		assert.NoError(t, err)
		assert.Equal(t, len(users), 2)
		session.Close()
	}
}

func storeTestNotifyAfterStore(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	storeLevelCallBack := []*ravendb.IMetadataDictionary{nil}
	sessionLevelCallback := []*ravendb.IMetadataDictionary{nil}

	fn := func(sender interface{}, event *ravendb.AfterSaveChangesEventArgs) {
		storeLevelCallBack[0] = event.GetDocumentMetadata()
	}
	store.AddAfterSaveChangesListener(fn)

	{
		session := openSessionMust(t, store)
		fn := func(sender interface{}, event *ravendb.AfterSaveChangesEventArgs) {
			sessionLevelCallback[0] = event.GetDocumentMetadata()
		}
		session.Advanced().AddAfterSaveChangesListener(fn)

		user1 := &User{}
		user1.setName("RavenDB")
		err = session.StoreWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		isLoaded := session.Advanced().IsLoaded("users/1")
		assert.True(t, isLoaded)

		changeVEctor, err := session.Advanced().GetChangeVectorFor(user1)
		assert.NoError(t, err)

		assert.NotNil(t, changeVEctor)
		session.Close()
	}

	assert.NotNil(t, storeLevelCallBack[0])
	assert.Equal(t, storeLevelCallBack[0], sessionLevelCallback[0])
	assert.NotNil(t, sessionLevelCallback[0])

	iMetadataDictionary := sessionLevelCallback[0]
	entrySet := iMetadataDictionary.EntrySet()
	for key, value := range entrySet {
		assert.NotEqual(t, key, "")
		assert.NotNil(t, value)
	}
}

func TestStore(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of java tests
	storeTestRefreshTest(t)
	storeTestStoreDocument(t)
	storeTestStoreDocuments(t)
	storeTestNotifyAfterStore(t)
}
