package tests

import (
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func deleteByQueryCanDeleteByQuery(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user1 := &User{
			Age: 5,
		}
		err = session.Store(user1)
		assert.NoError(t, err)

		user2 := &User{
			Age: 10,
		}
		err = session.Store(user2)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		indexQuery := ravendb.NewIndexQuery("from users where age == 5")
		operation, err := ravendb.NewDeleteByQueryOperation(indexQuery, nil)
		assert.NoError(t, err)
		asyncOp, err := store.Operations().SendAsync(operation, nil)
		assert.NoError(t, err)

		err = asyncOp.WaitForCompletion()
		assert.NoError(t, err)

		{
			session := openSessionMust(t, store)
			q := session.QueryCollectionForType(reflect.TypeOf(&User{}))
			count, err := q.Count()
			assert.NoError(t, err)
			assert.Equal(t, count, 1)
			session.Close()
		}
	}
}

func deleteByQueryCanDeleteByQueryWaitUsingChanges(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user1 := &User{
			Age: 5,
		}
		err = session.Store(user1)
		assert.NoError(t, err)

		user2 := &User{
			Age: 10,
		}
		err = session.Store(user2)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	var changesList chan *ravendb.OperationStatusChange
	var cancel ravendb.CancelFunc

	{
		changes := store.Changes("")
		err = changes.EnsureConnectedNow()
		require.NoError(t, err)

		changesList, cancel, err = changes.ForAllOperations()
		require.NoError(t, err)

		indexQuery := ravendb.NewIndexQuery("from users where age == 5")
		operation, err := ravendb.NewDeleteByQueryOperation(indexQuery, nil)
		assert.NoError(t, err)
		_, err = store.Operations().SendAsync(operation, nil)
		assert.NoError(t, err)

		select {
		case change := <-changesList:
			// ok, got a change
			expID := operation.Command.Result.OperationID
			assert.Equal(t, change.OperationID, expID)
			case <- time.After(15*time.Second):
				assert.Fail(t,"timed out waiting for operation change notification")
		}

		cancel()
		changes.Close()
	}
}

func TestDeleteByQuery(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	//TODO: match order of Java tests
	deleteByQueryCanDeleteByQuery(t, driver)

	deleteByQueryCanDeleteByQueryWaitUsingChanges(t, driver)
}
