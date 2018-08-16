package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

func storeTestRefreshTest(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user := NewUser()
		user.setName("RavenDB")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		{
			innerSession := openSessionMust(t, store)
			innerUserI, err := innerSession.Load(ravendb.GetTypeOf(&User{}), "users/1")
			innerUser := innerUserI.(*User)
			innerUser.setName("RavenDB 4.0")
			err = innerSession.SaveChanges()
			assert.NoError(t, err)
		}

		session.Advanced().Refresh(user)

		name := *user.GetName()
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
		user := NewUser()
		user.setName("RavenDB")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		userI, err := session.Load(ravendb.GetTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		user = userI.(*User)
		assert.NotNil(t, user)
		name := *user.GetName()
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
		user1 := NewUser()
		user1.setName("RavenDB")
		err = session.StoreWithID(user1, "users/1")
		assert.NoError(t, err)

		user2 := NewUser()
		user2.setName("Hibernating Rhinos")
		err = session.StoreWithID(user2, "users/2")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		users, err := session.LoadMulti(ravendb.GetTypeOf(&User{}), []string{"users/1", "users/2"})
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

		user1 := NewUser()
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
