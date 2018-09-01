package tests

import (
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func queriesWithCustomFunctions_queryCmpXchgWhere(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	err = store.Operations().Send(ravendb.NewPutCompareExchangeValueOperation("Tom", "Jerry", 0))
	assert.NoError(t, err)
	err = store.Operations().Send(ravendb.NewPutCompareExchangeValueOperation("Hera", "Zeus", 0))
	assert.NoError(t, err)
	err = store.Operations().Send(ravendb.NewPutCompareExchangeValueOperation("Gaya", "Uranus", 0))
	assert.NoError(t, err)
	err = store.Operations().Send(ravendb.NewPutCompareExchangeValueOperation("Jerry@gmail.com", "users/2", 0))
	assert.NoError(t, err)
	err = store.Operations().Send(ravendb.NewPutCompareExchangeValueOperation("Zeus@gmail.com", "users/1", 0))
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

		q := session.Advanced().DocumentQueryOld(ravendb.GetTypeOf(&User{}))
		q = q.WhereEquals("name", ravendb.CmpXchg_value("Hera"))
		q = q.WhereEquals("lastName", ravendb.CmpXchg_value("Tom"))

		query := q.GetIndexQuery().GetQuery()
		assert.Equal(t, query, "from Users where name = cmpxchg($p0) and lastName = cmpxchg($p1)")

		queryResult, err := q.ToListOld()
		assert.NoError(t, err)

		assert.Equal(t, len(queryResult), 1)

		user := queryResult[0].(*User)
		assert.Equal(t, *user.Name, "Zeus")

		q = session.Advanced().DocumentQueryOld(ravendb.GetTypeOf(&User{}))
		q = q.WhereNotEquals("name", ravendb.CmpXchg_value("Hera"))
		users, err := q.ToListOld()
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)

		user = users[0].(*User)
		assert.Equal(t, *user.Name, "Jerry")

		users, err = session.Advanced().RawQueryOld(ravendb.GetTypeOf(&User{}), "from Users where name = cmpxchg(\"Hera\")").ToListOld()
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)
		user = users[0].(*User)
		assert.Equal(t, *user.Name, "Zeus")

		session.Close()
	}
}

func TestQueriesWithCustomFunctions(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches the order of Java tests
	queriesWithCustomFunctions_queryCmpXchgWhere(t)
}
