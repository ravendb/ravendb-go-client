package tests

import (
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func queriesWithCustomFunctionsQueryCmpXchgWhere(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	op, err := ravendb.NewPutCompareExchangeValueOperation("Tom", "Jerry", 0)
	assert.NoError(t, err)
	err = store.Operations().Send(op, nil)
	assert.NoError(t, err)
	op, err = ravendb.NewPutCompareExchangeValueOperation("Hera", "Zeus", 0)
	assert.NoError(t, err)
	err = store.Operations().Send(op, nil)
	assert.NoError(t, err)
	op, err = ravendb.NewPutCompareExchangeValueOperation("Gaya", "Uranus", 0)
	assert.NoError(t, err)
	err = store.Operations().Send(op, nil)
	assert.NoError(t, err)
	op, err = ravendb.NewPutCompareExchangeValueOperation("Jerry@gmail.com", "users/2", 0)
	assert.NoError(t, err)
	err = store.Operations().Send(op, nil)
	assert.NoError(t, err)
	op, err = ravendb.NewPutCompareExchangeValueOperation("Zeus@gmail.com", "users/1", 0)
	assert.NoError(t, err)
	err = store.Operations().Send(op, nil)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		jerry := &User{}
		jerry.setName("Jerry")

		err = session.StoreWithID(jerry, "users/2")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		zeus := &User{}
		zeus.setName("Zeus")
		zeus.setLastName("Jerry")
		err = session.StoreWithID(zeus, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		q := session.Advanced().DocumentQuery()
		q = q.WhereEquals("name", ravendb.CmpXchgValue("Hera"))
		q = q.WhereEquals("lastName", ravendb.CmpXchgValue("Tom"))

		var users []*User
		err = q.GetResults(&users)
		assert.NoError(t, err)
		assert.Equal(t, len(users), 1)

		user := users[0]
		assert.Equal(t, *user.Name, "Zeus")

		query := q.GetIndexQuery().GetQuery()
		assert.Equal(t, query, "from Users where name = cmpxchg($p0) and lastName = cmpxchg($p1)")

		users = nil
		q = session.Advanced().DocumentQuery()
		q = q.WhereNotEquals("name", ravendb.CmpXchgValue("Hera"))
		err = q.GetResults(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)

		user = users[0]
		assert.Equal(t, *user.Name, "Jerry")

		users = nil
		{
			q, err := session.Advanced().RawQuery("from Users where name = cmpxchg(\"Hera\")")
			assert.NoError(t, err)
			err = q.GetResults(&users)
			assert.NoError(t, err)
		}

		assert.Equal(t, len(users), 1)
		user = users[0]
		assert.Equal(t, *user.Name, "Zeus")

		session.Close()
	}
}

func TestQueriesWithCustomFunctions(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches the order of Java tests
	queriesWithCustomFunctionsQueryCmpXchgWhere(t, driver)
}
