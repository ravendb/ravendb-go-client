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
			q := session.QueryType(reflect.TypeOf(&User{}))
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
	semaphore := make(chan bool)

	{
		changes := store.Changes("")
		err = changes.EnsureConnectedNow()
		require.NoError(t, err)

		action := func(v *ravendb.OperationStatusChange) {
			semaphore <- true
		}
		closer, err := changes.ForAllOperations(action)
		require.NoError(t, err)

		indexQuery := ravendb.NewIndexQuery("from users where age == 5")
		operation, err := ravendb.NewDeleteByQueryOperation(indexQuery, nil)
		assert.NoError(t, err)
		_, err = store.Operations().SendAsync(operation, nil)
		assert.NoError(t, err)

		timedOut := chanWaitTimedOut(semaphore, 15*time.Second)
		assert.False(t, timedOut)

		closer()
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
