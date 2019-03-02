package tests

import (
	"fmt"
	"testing"
	"time"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func changesTestSingleDocumentChangesCommon(t *testing.T, store *ravendb.DocumentStore) {

	changes := store.Changes("")
	err := changes.EnsureConnectedNow()
	assert.NoError(t, err)

	var cancel ravendb.CancelFunc
	{
		chDone := make(chan struct{})
		n := 0
		cb := func(documentChange *ravendb.DocumentChange) {
			if n == 0 {
				assert.NotNil(t, documentChange)
				assert.Equal(t, documentChange.ID, "users/1")
				assert.Equal(t, documentChange.Type, ravendb.DocumentChangePut)
				chDone <- struct{}{}
			} else {
				assert.Fail(t, "got too many (%d) changes")
			}
			n++
		}
		cancel, err = changes.ForDocument("users/1", cb)
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
		case <-chDone:
			// got a result
		case <-time.After(_reasonableWaitTime):
			assert.Fail(t, "timed out waiting for changesList to close")
		}
		cancel()
	}

	// at this point we should be unsubscribed from changes on 'users/1'
	{
		changesList := make(chan *ravendb.DocumentChange, 1)
		session := openSessionMust(t, store)
		user := &User{}
		user.setName("another name")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		select {
		case v := <-changesList:
			assert.Nil(t, v, "got too many changes")
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
		fmt.Printf("skipping changesTestChangesWithHttps() on windows")
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
		changes := store.Changes("")
		err = changes.EnsureConnectedNow()
		assert.NoError(t, err)

		{
			changesList := make(chan *ravendb.DocumentChange, 1)
			var unregister ravendb.CancelFunc
			cb := func(change *ravendb.DocumentChange) {
				changesList <- change
			}
			unregister, err = changes.ForAllDocuments(cb)
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
				assert.Fail(t,  "timed out waiting for changes")
			}

			select {
			case <-changesList:
				assert.Fail(t,  "got too many changes")
			case <-time.After(time.Second * 1):
				// ok, no changes
				assert.True(t, true)
			}

			unregister()
		}

		// at this point we should be unsubscribed from changes on 'users/1'

		{
			changesList := make(chan *ravendb.DocumentChange, 1)
			session := openSessionMust(t, store)
			user := &User{}
			user.setName("another name")
			err = session.StoreWithID(user, "users/1")
			assert.NoError(t, err)
			err = session.SaveChanges()
			assert.NoError(t, err)

			select {
			case v := <-changesList:
				assert.Nil(t, v, "got too many changes")
			case <-time.After(time.Second * 1):
				// ok, no changes
				assert.True(t, true)
			}
		}

		changes.Close()
		changes.Close() // call twice to make sure we're robust
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

	var cancel ravendb.CancelFunc

	{
		changesList := make(chan *ravendb.IndexChange, 1)
		changes := store.Changes("")
		err = changes.EnsureConnectedNow()
		assert.NoError(t, err)

		{
			cb := func(change *ravendb.IndexChange) {
				changesList <- change
			}
			cancel, err = changes.ForIndex(index.IndexName, cb)
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
				assert.Fail(t, "timed out waiting for changes")
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

	var cancel ravendb.CancelFunc

	{
		changesList := make(chan *ravendb.IndexChange, 1)
		changes := store.Changes("")
		err = changes.EnsureConnectedNow()
		assert.NoError(t, err)

		{
			cb := func(change *ravendb.IndexChange) {
				changesList <- change
			}
			cancel, err = changes.ForAllIndexes(cb)
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
				assert.Fail(t,  "timed out waiting for changes")
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

	var cancel ravendb.CancelFunc
	{
		changesList := make(chan *ravendb.DocumentChange, 1)
		changes := store.Changes("")
		err = changes.EnsureConnectedNow()
		assert.NoError(t, err)

		{
			cb := func(change *ravendb.DocumentChange) {
				changesList <- change
			}
			cancel, err = changes.ForDocumentsStartingWith("users/", cb)
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
				assert.Fail(t,  "timed out waiting for changes")
			}

			select {
			case documentChange := <-changesList:
				assert.Equal(t, documentChange.ID, "users/2")
			case <-time.After(time.Second * 2):
				assert.Fail(t,  "timed out waiting for changes")
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

	var cancel ravendb.CancelFunc
	{
		changesList := make(chan *ravendb.DocumentChange, 1)
		changes := store.Changes("")
		err = changes.EnsureConnectedNow()
		assert.NoError(t, err)

		{
			cb := func(change *ravendb.DocumentChange) {
				changesList <- change
			}
			cancel, err = changes.ForDocumentsInCollection("users", cb)
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
				assert.Fail(t,  "timed out waiting for changes")
			}

			select {
			case documentChange := <-changesList:
				assert.Equal(t, documentChange.ID, "users/2")
			case <-time.After(time.Second * 2):
				assert.Fail(t,  "timed out waiting for changes")
			}

			cancel()
		}
	}
}

func changesTestNotificationOnWrongDatabaseShouldNotCrashServer(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	closer := disableLogFailedRequests()
	defer closer()

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
