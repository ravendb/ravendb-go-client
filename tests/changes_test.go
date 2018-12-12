package tests

import (
	"strconv"
	"testing"
	"time"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func changesTest_singleDocumentChangesCommon(t *testing.T, store *ravendb.DocumentStore) {
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
			assert.Equal(t, documentChange.Type, ravendb.DocumentChangeTypes_PUT)

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

func changesTest_singleDocumentChanges(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()
	changesTest_singleDocumentChangesCommon(t, store)
}

func changesTest_changesWithHttps(t *testing.T) {
	if isWindows() {
		t.Skip("skipping https test on windows")
		return
	}
	store := getSecuredDocumentStoreMust(t)
	defer store.Close()
	changesTest_singleDocumentChangesCommon(t, store)
}

func changesTest_allDocumentsChanges(t *testing.T) {

	var err error
	store := getDocumentStoreMust(t)
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
				assert.Equal(t, documentChange.Type, ravendb.DocumentChangeTypes_PUT)

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

func changesTest_singleIndexChanges(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
			operation := ravendb.NewSetIndexesPriorityOperation(index.IndexName, ravendb.IndexPriority_LOW)
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

func changesTest_allIndexChanges(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
			operation := ravendb.NewSetIndexesPriorityOperation(index.IndexName, ravendb.IndexPriority_LOW)
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

func changesTest_notificationOnWrongDatabase_ShouldNotCrashServer(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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

func changesTest_resourcesCleanup(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
					assert.Equal(t, documentChange.Type, ravendb.DocumentChangeTypes_PUT)

				case <-time.After(time.Second * 10):
					assert.True(t, false, "timed out waiting for changes")
				}

				subscription.Close()
			}

		}
	}
}

func TestChanges(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// follows execution order of java tests
	changesTest_allDocumentsChanges(t)
	changesTest_singleDocumentChanges(t)
	changesTest_resourcesCleanup(t)
	changesTest_changesWithHttps(t)
	changesTest_singleIndexChanges(t)
	changesTest_notificationOnWrongDatabase_ShouldNotCrashServer(t)
	changesTest_allIndexChanges(t)
}
