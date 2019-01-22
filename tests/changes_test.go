package tests

import (
	"strconv"
	"testing"
	"time"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func changesTestSingleDocumentChangesCommon(t *testing.T, store *ravendb.DocumentStore) {
	changesList := make(chan *ravendb.DocumentChange, 8)

	changes := store.Changes()
	err := changes.EnsureConnectedNow()
	assert.NoError(t, err)

	observable, err := changes.ForDocument("users/1")
	assert.NoError(t, err)

	{
		action := func(v interface{}) {
			change := v.(*ravendb.DocumentChange)
			changesList <- change
		}
		observer := ravendb.NewActionBasedObserver(action)
		subscription := observable.Subscribe(observer)

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
		subscription.Close()
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

		changes := store.Changes()
		err = changes.EnsureConnectedNow()
		assert.NoError(t, err)

		observable, err := changes.ForAllDocuments()
		assert.NoError(t, err)

		{
			action := func(v interface{}) {
				change := v.(*ravendb.DocumentChange)
				changesList <- change
			}
			observer := ravendb.NewActionBasedObserver(action)
			subscription := observable.Subscribe(observer)

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
			subscription.Close()
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
	}
}

// Note: UsersByName is the same as makeUsersByNameIndex in query_test.go

func changesTestSingleIndexChanges(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := makeUsersByNameIndex()
	err = store.ExecuteIndex(index)
	assert.NoError(t, err)

	changesList := make(chan *ravendb.IndexChange, 8)

	{
		changes := store.Changes()
		err = changes.EnsureConnectedNow()
		assert.NoError(t, err)

		observable, err := changes.ForIndex(index.IndexName)
		assert.NoError(t, err)

		{
			action := func(v interface{}) {
				change := v.(*ravendb.IndexChange)
				changesList <- change
			}
			observer := ravendb.NewActionBasedObserver(action)
			subscription := observable.Subscribe(observer)
			time.Sleep(500 * time.Millisecond)
			//SetIndexesPriorityOperation
			operation := ravendb.NewSetIndexesPriorityOperation(index.IndexName, ravendb.IndexPriorityLow)
			err = store.Maintenance().Send(operation)
			assert.NoError(t, err)

			select {
			case indexChange := <-changesList:
				assert.Equal(t, indexChange.Name, index.IndexName)
			case <-time.After(time.Second * 2):
				assert.True(t, false, "timed out waiting for changes")
			}
			subscription.Close()
		}
	}
}

func changesTestAllIndexChanges(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := makeUsersByNameIndex()
	err = store.ExecuteIndex(index)
	assert.NoError(t, err)

	changesList := make(chan *ravendb.IndexChange, 8)

	{
		changes := store.Changes()
		err = changes.EnsureConnectedNow()
		assert.NoError(t, err)

		observable, err := changes.ForAllIndexes()
		assert.NoError(t, err)

		{
			action := func(v interface{}) {
				change := v.(*ravendb.IndexChange)
				changesList <- change
			}
			observer := ravendb.NewActionBasedObserver(action)
			subscription := observable.Subscribe(observer)
			time.Sleep(500 * time.Millisecond)
			operation := ravendb.NewSetIndexesPriorityOperation(index.IndexName, ravendb.IndexPriorityLow)
			err = store.Maintenance().Send(operation)
			assert.NoError(t, err)

			select {
			case indexChange := <-changesList:
				assert.Equal(t, indexChange.Name, index.IndexName)
			case <-time.After(time.Second * 2):
				assert.True(t, false, "timed out waiting for changes")
			}
			subscription.Close()
		}
	}
}

func changesTestNotificationOnWrongDatabaseShouldNotCrashServer(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	semaphore := make(chan bool, 1)
	semaphore <- true // acquire

	changes := store.ChangesWithDatabaseName("no_such_db")

	onError := func(e error) {
		<-semaphore // release
	}
	changes.AddOnError(onError)

	select {
	case <-semaphore:
		// do nothing
	case <-time.After(time.Second * 15):
		assert.True(t, false, "timed out waiting for error")
	}

	op := ravendb.NewGetStatisticsOperation()
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)
}

func changesTestResourcesCleanup(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := makeUsersByNameIndex()
	err = store.ExecuteIndex(index)
	assert.NoError(t, err)

	// repeat this few times and watch deadlocks
	for i := 0; i < 100; i++ {
		changesList := make(chan *ravendb.DocumentChange, 8)

		{
			changes := store.Changes()
			err = changes.EnsureConnectedNow()
			assert.NoError(t, err)

			observable, err := changes.ForDocument("users/" + strconv.Itoa(i))
			assert.NoError(t, err)

			{
				action := func(v interface{}) {
					change := v.(*ravendb.DocumentChange)
					changesList <- change
				}
				observer := ravendb.NewActionBasedObserver(action)
				subscription := observable.Subscribe(observer)

				{
					session := openSessionMust(t, store)
					user := &User{}
					err = session.StoreWithID(user, "users/"+strconv.Itoa(i))
					assert.NoError(t, err)
					err = session.SaveChanges()
					assert.NoError(t, err)
				}

				select {
				case documentChange := <-changesList:
					assert.NotNil(t, documentChange)
					assert.Equal(t, documentChange.ID, "users/"+strconv.Itoa(i))
					assert.Equal(t, documentChange.Type, ravendb.DocumentChangePut)

				case <-time.After(time.Second * 10):
					assert.True(t, false, "timed out waiting for changes")
				}

				subscription.Close()
			}

		}
	}
}

func TestChanges(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// follows execution order of java tests
	changesTestAllDocumentsChanges(t, driver)
	changesTestSingleDocumentChanges(t, driver)
	changesTestResourcesCleanup(t, driver)
	changesTestChangesWithHttps(t, driver)
	changesTestSingleIndexChanges(t, driver)
	changesTestNotificationOnWrongDatabaseShouldNotCrashServer(t, driver)
	changesTestAllIndexChanges(t, driver)
}
