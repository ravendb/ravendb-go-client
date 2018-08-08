package ravendb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func storeTestRefreshTest(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user := NewUser()
		user.setName("RavenDB")
		err = session.StoreEntityWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		{
			innerSession := openSessionMust(t, store)
			innerUserI, err := innerSession.load(getTypeOf(&User{}), "users/1")
			innerUser := innerUserI.(*User)
			innerUser.setName("RavenDB 4.0")
			err = innerSession.SaveChanges()
			assert.NoError(t, err)
		}

		session.advanced().refresh(user)

		name := *user.getName()
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
		err = session.StoreEntityWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		userI, err := session.load(getTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		user = userI.(*User)
		assert.NotNil(t, user)
		name := *user.getName()
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
		err = session.StoreEntityWithID(user1, "users/1")
		assert.NoError(t, err)

		user2 := NewUser()
		user2.setName("Hibernating Rhinos")
		err = session.StoreEntityWithID(user2, "users/2")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		users, err := session.loadMulti(getTypeOf(&User{}), []string{"users/1", "users/2"})
		assert.NoError(t, err)
		assert.Equal(t, len(users), 2)
		session.Close()
	}
}

func storeTestNotifyAfterStore(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	storeLevelCallBack := []*IMetadataDictionary{nil}
	sessionLevelCallback := []*IMetadataDictionary{nil}

	fn := func(sender interface{}, event *AfterSaveChangesEventArgs) {
		storeLevelCallBack[0] = event.getDocumentMetadata()
	}
	store.addAfterSaveChangesListener(fn)

	{
		session := openSessionMust(t, store)
		fn := func(sender interface{}, event *AfterSaveChangesEventArgs) {
			sessionLevelCallback[0] = event.getDocumentMetadata()
		}
		session.advanced().addAfterSaveChangesListener(fn)

		user1 := NewUser()
		user1.setName("RavenDB")
		err = session.StoreEntityWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		isLoaded := session.advanced().IsLoaded("users/1")
		assert.True(t, isLoaded)

		changeVEctor, err := session.advanced().getChangeVectorFor(user1)
		assert.NoError(t, err)

		assert.NotNil(t, changeVEctor)
		session.Close()
	}

	assert.NotNil(t, storeLevelCallBack[0])
	assert.Equal(t, storeLevelCallBack[0], sessionLevelCallback[0])
	assert.NotNil(t, sessionLevelCallback[0])

	iMetadataDictionary := sessionLevelCallback[0]
	entrySet := iMetadataDictionary.entrySet()
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
	defer func() {
		r := recover()
		destroyDriver()
		if r != nil {
			panic(r)
		}
	}()

	// matches order of java tests
	storeTestRefreshTest(t)
	storeTestStoreDocument(t)
	storeTestStoreDocuments(t)
	storeTestNotifyAfterStore(t)
}
