package tests

import (
	"testing"
	"time"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func changesTest_singleDocumentChanges(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		changesList := make(chan *ravendb.DocumentChange, 8)

		changes := store.Changes()
		err = changes.EnsureConnectedNow()
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
}

func changesTest_changesWithHttps(t *testing.T) {}

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

func changesTest_singleIndexChanges(t *testing.T) {}

func changesTest_allIndexChanges(t *testing.T) {}

func changesTest_notificationOnWrongDatabase_ShouldNotCrashServer(t *testing.T) {}

func changesTest_resourcesCleanup(t *testing.T) {}

/*
   public static class UsersByName extends AbstractIndexCreationTask {
       public UsersByName() {

           map = "from c in docs.Users select new " +
                   " {" +
                   "    c.name, " +
                   "    count = 1" +
                   "}";

           reduce = "from result in results " +
                   "group result by result.name " +
                   "into g " +
                   "select new " +
                   "{ " +
                   "  name = g.Key, " +
                   "  count = g.Sum(x => x.count) " +
                   "}";
       }
   }
*/

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
