package ravendb

import (
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

func loadTest_canDeleteByQuery(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user1 := NewUser()
		user1.setAge(5)
		err = session.StoreEntity(user1)
		assert.NoError(t, err)

		user2 := NewUser()
		user2.setAge(10)
		err = session.StoreEntity(user2)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	// TODO: fix loadTest_canDeleteByQuery
	if false {
		{
			indexQuery := NewIndexQuery("from users where age == 5")
			operation := NewDeleteByQueryOperation(indexQuery)
			asyncOp, err := store.operations().sendAsync(operation)
			assert.NoError(t, err)

			err = asyncOp.waitForCompletion()
			assert.NoError(t, err)

			{
				session := openSessionMust(t, store)
				q := session.query(getTypeOf(&User{}))
				count, err := q.count()
				assert.NoError(t, err)
				assert.Equal(t, count, 1)
				session.Close()
			}
		}
	}
}

func TestDeleteByQuery(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	if useProxy() {
		proxy.ChangeLogFile("trace_delete_by_query_go.txt")
	}

	createTestDriver()
	defer deleteTestDriver()

	// matches order of Java tests
	loadTest_canDeleteByQuery(t)
}
