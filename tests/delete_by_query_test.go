package tests

import (
	"reflect"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func loadTest_canDeleteByQuery(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user1 := &User{}
		user1.Age = 5
		err = session.Store(user1)
		assert.NoError(t, err)

		user2 := &User{}
		user2.Age = 10
		err = session.Store(user2)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		indexQuery := ravendb.NewIndexQuery("from users where age == 5")
		operation := ravendb.NewDeleteByQueryOperation(indexQuery)
		asyncOp, err := store.Operations().SendAsync(operation)
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

func TestDeleteByQuery(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	loadTest_canDeleteByQuery(t, driver)
}
