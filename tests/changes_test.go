package tests

import (
	"testing"
	"time"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func changesTestSingleDocumentChangesCommon(t *testing.T, store *ravendb.DocumentStore) {
	changesList := make(chan *ravendb.DocumentChange, 8)

	changes := store.Changes("")
	err := changes.EnsureConnectedNow()
	assert.NoError(t, err)

	{
		action := func(change *ravendb.DocumentChange) {
			changesList <- change
		}
		cancel, err := changes.ForDocument("users/1", action)
		assert.NoError(t, err)

		{
			session := openSessionMust(t, store)
			user := &User{}
			err = session.StoreWithID(user, "users/1")
			assert.NoError(t, err)
			err = session.SaveChanges()
			assert.NoError(t, err)
		}

		select {
		case documentChange := <-changesList:
			assert.NotNil(t, documentChange)
			assert.Equal(t, documentChange.ID, "users/1")
			assert.Equal(t, documentChange.Type, ravendb.DocumentChangePut)

		case <-time.After(time.Second * 2):
			assert.True(t, false, "timed out waiting for changes")
		}

		select {
		case <-changesList:
			assert.True(t, false, "got too many changes")
		case <-time.After(time.Second * 1):
			// ok, no changes
			assert.True(t, true)
		}

		cancel()
	}

	// at this point we should be unsubscribed from changes on 'users/1'
	{
		session := openSessionMust(t, store)
		user := &User{}
		user.setName("another name")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		select {
		case <-changesList:
			assert.True(t, false, "got too many changes")
		case <-time.After(time.Second * 1):
			// ok, no changes
			assert.True(t, true)
		}
	}

	changes.Close()
}

func changesTestSingleDocumentChanges(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()
	changesTestSingleDocumentChangesCommon(t, store)
}

func changesTestChangesWithHttps(t *testing.T, driver *RavenTestDriver) {
	if isWindows() {
		t.Skip("skipping https test on windows")
		return
	}
	store := driver.getSecuredDocumentStoreMust(t)
	defer store.Close()
	changesTestSingleDocumentChangesCommon(t, store)
}

func changesTestAllDocumentsChanges(t *testing.T, driver *RavenTestDriver) {

	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		changesList := make(chan *ravendb.DocumentChange, 8)

		changes := store.Changes("")
		err = changes.EnsureConnectedNow()
		assert.NoError(t, err)

		{
			action := func(change *ravendb.DocumentChange) {
				changesList <- change
			}
			cancel, err := changes.ForAllDocuments(action)
			assert.NoError(t, err)

			{
				session := openSessionMust(t, store)
				user := &User{}
				err = session.StoreWithID(user, "users/1")
				assert.NoError(t, err)
				err = session.SaveChanges()
				assert.NoError(t, err)
			}

			select {
			case documentChange := <-changesList:
				assert.NotNil(t, documentChange)
				assert.Equal(t, documentChange.ID, "users/1")
				assert.Equal(t, documentChange.Type, ravendb.DocumentChangePut)

			case <-time.After(time.Second * 2):
				assert.True(t, false, "timed out waiting for changes")
			}

			select {
			case <-changesList:
				assert.True(t, false, "got too many changes")
			case <-time.After(time.Second * 1):
				// ok, no changes
				assert.True(t, true)
			}

			cancel()
		}

		// at this point we should be unsubscribed from changes on 'users/1'

		{
			session := openSessionMust(t, store)
			user := &User{}
			user.setName("another name")
			err = session.StoreWithID(user, "users/1")
			assert.NoError(t, err)
			err = session.SaveChanges()
			assert.NoError(t, err)

			select {
			case <-changesList:
				assert.True(t, false, "got too many changes")
			case <-time.After(time.Second * 1):
				// ok, no changes
				assert.True(t, true)
			}
		}

		changes.Close()
	}
}

// Note: UsersByName is the same as makeUsersByNameIndex in query_test.go

func changesTestSingleIndexChanges(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := makeUsersByNameIndex()
	err = store.ExecuteIndex(index, "")
	assert.NoError(t, err)

	changesList := make(chan *ravendb.IndexChange, 8)

	{
		changes := store.Changes("")
		err = changes.EnsureConnectedNow()
		assert.NoError(t, err)

		{
			action := func(change *ravendb.IndexChange) {
				changesList <- change
			}
			cancel, err := changes.ForIndex(index.IndexName, action)
			assert.NoError(t, err)

			time.Sleep(500 * time.Millisecond)
			operation, err := ravendb.NewSetIndexesPriorityOperation(index.IndexName, ravendb.IndexPriorityLow)
			assert.NoError(t, err)
			err = store.Maintenance().Send(operation)
			assert.NoError(t, err)

			select {
			case indexChange := <-changesList:
				assert.Equal(t, indexChange.Name, index.IndexName)
			case <-time.After(time.Second * 2):
				assert.True(t, false, "timed out waiting for changes")
			}

			cancel()
		}

		changes.Close()
	}
}

func changesTestAllIndexChanges(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := makeUsersByNameIndex()
	err = store.ExecuteIndex(index, "")
	assert.NoError(t, err)

	changesList := make(chan *ravendb.IndexChange, 8)

	{
		changes := store.Changes("")
		err = changes.EnsureConnectedNow()
		assert.NoError(t, err)

		{
			action := func(change *ravendb.IndexChange) {
				changesList <- change
			}
			cancel, err := changes.ForAllIndexes(action)
			assert.NoError(t, err)

			time.Sleep(500 * time.Millisecond)
			operation, err := ravendb.NewSetIndexesPriorityOperation(index.IndexName, ravendb.IndexPriorityLow)
			assert.NoError(t, err)
			err = store.Maintenance().Send(operation)
			assert.NoError(t, err)

			select {
			case indexChange := <-changesList:
				assert.Equal(t, indexChange.Name, index.IndexName)
			case <-time.After(time.Second * 2):
				assert.True(t, false, "timed out waiting for changes")
			}
			cancel()
		}

		changes.Close()
	}
}

func changesTestCanCanNotificationAboutDocumentsStartingWiths(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	changesList := make(chan *ravendb.DocumentChange)

	{
		changes := store.Changes("")
		err = changes.EnsureConnectedNow()
		assert.NoError(t, err)

		{
			action := func(change *ravendb.DocumentChange) {
				changesList <- change
			}
			cancel, err := changes.ForDocumentsStartingWith("users/", action)
			assert.NoError(t, err)

			{
				session := openSessionMust(t, store)
				err = session.StoreWithID(&User{}, "users/1")
				assert.NoError(t, err)
				err = session.SaveChanges()
				assert.NoError(t, err)

				session.Close()
			}

			{
				session := openSessionMust(t, store)
				err = session.StoreWithID(&User{}, "differentDocumentPrefix/1")
				assert.NoError(t, err)
				err = session.SaveChanges()
				assert.NoError(t, err)

				session.Close()
			}

			{
				session := openSessionMust(t, store)
				err = session.StoreWithID(&User{}, "users/2")
				assert.NoError(t, err)
				err = session.SaveChanges()
				assert.NoError(t, err)

				session.Close()
			}

			select {
			case documentChange := <-changesList:
				assert.Equal(t, documentChange.ID, "users/1")
			case <-time.After(time.Second * 2):
				assert.True(t, false, "timed out waiting for changes")
			}

			select {
			case documentChange := <-changesList:
				assert.Equal(t, documentChange.ID, "users/2")
			case <-time.After(time.Second * 2):
				assert.True(t, false, "timed out waiting for changes")
			}

			cancel()
		}

		changes.Close()
	}
}

func changesTestCanCanNotificationAboutDocumentsFromCollection(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	changesList := make(chan *ravendb.DocumentChange)

	{
		changes := store.Changes("")
		err = changes.EnsureConnectedNow()
		assert.NoError(t, err)

		{
			action := func(change *ravendb.DocumentChange) {
				changesList <- change
			}
			cancel, err := changes.ForDocumentsInCollection("users", action)
			assert.NoError(t, err)

			{
				session := openSessionMust(t, store)
				err = session.StoreWithID(&User{}, "users/1")
				assert.NoError(t, err)
				err = session.SaveChanges()
				assert.NoError(t, err)

				session.Close()
			}

			{
				session := openSessionMust(t, store)
				err = session.StoreWithID(&Order{}, "orders/1")
				assert.NoError(t, err)
				err = session.SaveChanges()
				assert.NoError(t, err)

				session.Close()
			}

			{
				session := openSessionMust(t, store)
				err = session.StoreWithID(&User{}, "users/2")
				assert.NoError(t, err)
				err = session.SaveChanges()
				assert.NoError(t, err)

				session.Close()
			}

			select {
			case documentChange := <-changesList:
				assert.Equal(t, documentChange.ID, "users/1")
			case <-time.After(time.Second * 2):
				assert.True(t, false, "timed out waiting for changes")
			}

			select {
			case documentChange := <-changesList:
				assert.Equal(t, documentChange.ID, "users/2")
			case <-time.After(time.Second * 2):
				assert.True(t, false, "timed out waiting for changes")
			}

			cancel()
		}
	}
}

func changesTestNotificationOnWrongDatabaseShouldNotCrashServer(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	changes := store.Changes("no_such_db")
	err = changes.EnsureConnectedNow()
	assert.NotNil(t, err)
	_, ok := err.(*ravendb.DatabaseDoesNotExistError)
	assert.True(t, ok)

	op := ravendb.NewGetStatisticsOperation("")
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)
}

func TestChanges(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// follows execution order of java tests
	changesTestAllDocumentsChanges(t, driver)
	changesTestSingleDocumentChanges(t, driver)
	changesTestChangesWithHttps(t, driver)
	changesTestSingleIndexChanges(t, driver)
	changesTestNotificationOnWrongDatabaseShouldNotCrashServer(t, driver)
	changesTestAllIndexChanges(t, driver)

	// TODO: order different than Java's
	changesTestCanCanNotificationAboutDocumentsStartingWiths(t, driver)
	changesTestCanCanNotificationAboutDocumentsFromCollection(t, driver)
}
