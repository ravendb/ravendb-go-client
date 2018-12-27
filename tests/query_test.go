package tests

import (
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func query_querySimple(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user1 := &User{}
		user1.setName("John")

		user2 := &User{}
		user2.setName("Jane")

		user3 := &User{}
		user3.setName("Tarzan")

		err = session.StoreWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.StoreWithID(user2, "users/2")
		assert.NoError(t, err)
		err = session.StoreWithID(user3, "users/3")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		var queryResult []*User
		q := session.Advanced().DocumentQueryAllOld(reflect.TypeOf(&User{}), "", "users", false)
		err := q.ToList(&queryResult)
		assert.NoError(t, err)
		assert.Equal(t, len(queryResult), 3)

		session.Close()
	}
}

func query_queryLazily(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user1 := &User{}
		user1.setName("John")

		user2 := &User{}
		user2.setName("Jane")

		user3 := &User{}
		user3.setName("Tarzan")

		err = session.StoreWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.StoreWithID(user2, "users/2")
		assert.NoError(t, err)
		err = session.StoreWithID(user3, "users/3")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		lazyQuery := session.QueryOld(reflect.TypeOf(&User{})).Lazily()
		queryResultI, err := lazyQuery.GetValue()
		assert.NoError(t, err)
		queryResult := queryResultI.([]*User)
		assert.Equal(t, 3, len(queryResult))

		assert.Equal(t, *queryResult[0].Name, "John")
		assert.Equal(t, *queryResult[1].Name, "Jane")
		assert.Equal(t, *queryResult[2].Name, "Tarzan")
	}
}

func query_collectionsStats(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user1 := &User{}
		user1.setName("John")

		user2 := &User{}
		user2.setName("Jane")

		err = session.StoreWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.StoreWithID(user2, "users/2")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	op := ravendb.NewGetCollectionStatisticsOperation()
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)
	stats := op.Command.Result
	assert.Equal(t, stats.CountOfDocuments, 2)
	coll := stats.Collections["Users"]
	assert.Equal(t, coll, 2)
}

func query_queryWithWhereClause(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user1 := &User{}
		user1.setName("John")

		user2 := &User{}
		user2.setName("Jane")

		user3 := &User{}
		user3.setName("Tarzan")

		err = session.StoreWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.StoreWithID(user2, "users/2")
		assert.NoError(t, err)
		err = session.StoreWithID(user3, "users/3")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		var queryResult []*User
		q := session.QueryWithQueryOld(reflect.TypeOf(&User{}), ravendb.Query_collection("users"))
		q = q.WhereStartsWith("name", "J")
		err := q.ToList(&queryResult)
		assert.NoError(t, err)

		var queryResult2 []*User
		q2 := session.QueryWithQueryOld(reflect.TypeOf(&User{}), ravendb.Query_collection("users"))
		q2 = q2.WhereEquals("name", "Tarzan")
		err = q2.ToList(&queryResult2)
		assert.NoError(t, err)

		var queryResult3 []*User
		q3 := session.QueryWithQueryOld(reflect.TypeOf(&User{}), ravendb.Query_collection("users"))
		q3 = q3.WhereEndsWith("name", "n")
		err = q3.ToList(&queryResult3)
		assert.NoError(t, err)

		assert.Equal(t, len(queryResult), 2)
		assert.Equal(t, len(queryResult2), 1)
		assert.Equal(t, len(queryResult3), 2)

		session.Close()
	}
}

func query_queryMapReduceWithCount(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var results []*ReduceResult
		q := session.QueryOld(reflect.TypeOf(&User{}))
		q2 := q.GroupBy("name")
		q2 = q2.SelectKey()
		q = q2.SelectCount()
		q = q.OrderByDescending("count")
		// TODO: OfType is not needed. See https://github.com/ravendb/ravendb-go-client/issues/95
		q = q.OfType(reflect.TypeOf(&ReduceResult{}))
		err := q.ToList(&results)
		assert.NoError(t, err)

		{
			result := results[0]
			assert.Equal(t, result.Count, 2)
			assert.Equal(t, result.Name, "John")
		}

		{
			result := results[1]
			assert.Equal(t, result.Count, 1)
			assert.Equal(t, result.Name, "Tarzan")
		}

		session.Close()
	}
}

func query_queryMapReduceWithSum(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var results []*ReduceResult
		q := session.QueryType(reflect.TypeOf(&User{}))
		q2 := q.GroupBy("name")
		q2 = q2.SelectKey()
		f := &ravendb.GroupByField{
			FieldName: "age",
		}
		q = q2.SelectSum(f)
		q = q.OrderByDescending("age")
		q = q.OfType(reflect.TypeOf(&ReduceResult{}))
		err := q.ToList(&results)
		assert.NoError(t, err)

		{
			result := results[0]
			assert.Equal(t, result.Age, 8)
			assert.Equal(t, result.Name, "John")
		}

		{
			result := results[1]
			assert.Equal(t, result.Age, 2)
			assert.Equal(t, result.Name, "Tarzan")
		}

		session.Close()
	}
}

func query_queryMapReduceIndex(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var results []*ReduceResult
		q := session.QueryWithQueryOld(reflect.TypeOf(&ReduceResult{}), ravendb.Query_index("UsersByName"))
		q = q.OrderByDescending("count")
		err := q.ToList(&results)
		assert.NoError(t, err)

		{
			result := results[0]
			assert.Equal(t, result.Count, 2)
			assert.Equal(t, result.Name, "John")
		}

		{
			result := results[1]
			assert.Equal(t, result.Count, 1)
			assert.Equal(t, result.Name, "Tarzan")
		}

		session.Close()
	}
}

func query_querySingleProperty(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		q := session.QueryType(reflect.TypeOf(&User{}))
		q = q.AddOrderWithOrdering("age", true, ravendb.OrderingType_LONG)
		q = q.SelectFields(reflect.TypeOf(int(0)), "age")
		ages, err := q.ToListOld()
		assert.NoError(t, err)

		assert.Equal(t, len(ages), 3)

		for _, n := range []int{5, 3, 2} {
			assert.True(t, ravendb.InterfaceArrayContains(ages, n))
		}

		session.Close()
	}
}

func query_queryWithSelect(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var usersAge []*User
		q := session.QueryOld(reflect.TypeOf(&User{}))
		q = q.SelectFields(reflect.TypeOf(&User{}), "age")
		err := q.ToList(&usersAge)
		assert.NoError(t, err)

		for _, user := range usersAge {
			assert.True(t, user.Age >= 0)
			assert.NotEmpty(t, user.ID)
		}

		session.Close()
	}
}

func query_queryWithWhereIn(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.Query()
		q = q.WhereIn("name", []interface{}{"Tarzan", "no_such"})
		err := q.ToList(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)

		session.Close()
	}
}

func query_queryWithWhereBetween(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.Query()
		q = q.WhereBetween("age", 4, 5)
		err := q.ToList(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)

		user := users[0]
		assert.Equal(t, *user.Name, "John")

		session.Close()
	}
}

func query_queryWithWhereLessThan(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.Query()
		q = q.WhereLessThan("age", 3)
		err := q.ToList(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)

		user := users[0]
		assert.Equal(t, *user.Name, "Tarzan")

		session.Close()
	}
}

func query_queryWithWhereLessThanOrEqual(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.Query()
		q = q.WhereLessThanOrEqual("age", 3)
		err := q.ToList(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 2)

		session.Close()
	}
}

func query_queryWithWhereGreaterThan(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.Query()
		q = q.WhereGreaterThan("age", 3)
		err := q.ToList(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)

		user := users[0]
		assert.Equal(t, *user.Name, "John")

		session.Close()
	}
}

func query_queryWithWhereGreaterThanOrEqual(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.Query()
		q = q.WhereGreaterThanOrEqual("age", 3)
		err := q.ToList(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 2)

		session.Close()
	}
}

type UserProjection struct {
	ID   string
	Name string
}

func query_queryWithProjection(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		q := session.QueryType(reflect.TypeOf(&User{}))
		q = q.SelectFields(reflect.TypeOf(&UserProjection{}))
		var projections []*UserProjection
		err := q.ToList(&projections)
		assert.NoError(t, err)

		assert.Equal(t, len(projections), 3)

		for _, projection := range projections {
			assert.NotEmpty(t, projection.ID)

			assert.NotEmpty(t, projection.Name)
		}

		session.Close()
	}
}

func query_queryWithProjection2(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		q := session.QueryType(reflect.TypeOf(&User{}))
		q = q.SelectFields(reflect.TypeOf(&UserProjection{}), "lastName")
		var projections []*UserProjection
		err := q.ToList(&projections)
		assert.NoError(t, err)

		assert.Equal(t, len(projections), 3)

		for _, projection := range projections {
			assert.NotEmpty(t, projection.ID)

			assert.Empty(t, projection.Name) // we didn't specify this field in mapping
		}

		session.Close()
	}
}

func query_queryDistinct(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		q := session.QueryType(reflect.TypeOf(&User{}))
		q = q.SelectFields(reflect.TypeOf(""), "name")
		q = q.Distinct()
		uniqueNames, err := q.ToListOld()
		assert.NoError(t, err)

		assert.Equal(t, len(uniqueNames), 2)
		assert.True(t, ravendb.InterfaceArrayContains(uniqueNames, "Tarzan"))
		assert.True(t, ravendb.InterfaceArrayContains(uniqueNames, "John"))

		session.Close()
	}
}

func query_querySearchWithOr(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var uniqueNames []*User
		q := session.Query()
		q = q.SearchWithOperator("name", "Tarzan John", ravendb.SearchOperator_OR)
		err := q.ToList(&uniqueNames)
		assert.NoError(t, err)

		assert.Equal(t, len(uniqueNames), 3)

		session.Close()
	}
}

func query_queryNoTracking(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.Query()
		q = q.NoTracking()
		err := q.ToList(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 3)

		for _, user := range users {
			isLoaded := session.IsLoaded(user.ID)
			assert.False(t, isLoaded)
		}

		session.Close()
	}
}

func query_querySkipTake(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.Query()
		q = q.OrderBy("name")
		q = q.Skip(2)
		q = q.Take(1)
		err := q.ToList(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)

		user := users[0]
		assert.Equal(t, *user.Name, "Tarzan")

		session.Close()
	}
}

func query_rawQuerySkipTake(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.RawQuery("from users")
		q = q.Skip(2)
		q = q.Take(1)
		err = q.ToList(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)
		user := users[0]
		assert.Equal(t, *user.Name, "Tarzan")

		session.Close()
	}
}

func query_parametersInRawQuery(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.RawQuery("from users where age == $p0")
		q = q.AddParameter("p0", 5)
		err = q.ToList(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)
		user := users[0]
		assert.Equal(t, *user.Name, "John")

		session.Close()
	}
}

func query_queryLucene(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.Query()
		q = q.WhereLucene("name", "Tarzan")
		err := q.ToList(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)

		for _, user := range users {
			assert.Equal(t, *user.Name, "Tarzan")
		}

		session.Close()
	}
}

func query_queryWhereExact(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		{
			var users []*User
			q := session.Query()
			q = q.WhereEquals("name", "tarzan")
			err := q.ToList(&users)
			assert.NoError(t, err)

			assert.Equal(t, len(users), 1)
		}

		{
			var users []*User
			q := session.Query()
			q = q.WhereEquals("name", "tarzan").Exact()
			err := q.ToList(&users)
			assert.NoError(t, err)

			assert.Equal(t, len(users), 0) // we queried for tarzan with exact
		}

		{
			var users []*User
			q := session.Query()
			q = q.WhereEquals("name", "Tarzan").Exact()
			err := q.ToList(&users)
			assert.NoError(t, err)

			assert.Equal(t, len(users), 1) // we queried for Tarzan with exact
		}

		session.Close()
	}
}

func query_queryWhereNot(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		{
			var res []*User
			q := session.Query()
			q = q.Not()
			q = q.WhereEquals("name", "tarzan")
			err := q.ToList(&res)

			assert.NoError(t, err)

			assert.Equal(t, len(res), 2)
		}

		{
			var res []*User
			q := session.Query()
			q = q.WhereNotEquals("name", "tarzan")
			err := q.ToList(&res)

			assert.NoError(t, err)

			assert.Equal(t, len(res), 2)
		}

		{
			var res []*User
			q := session.Query()
			q = q.WhereNotEquals("name", "Tarzan").Exact()
			err := q.ToList(&res)

			assert.NoError(t, err)

			assert.Equal(t, len(res), 2)
		}

		session.Close()
	}
}

/*
TODO: is this used?
 static class Result {
	 long delay

	 long getDelay() {
		return delay
	}

	  setDelay(long delay) {
		this.delay = delay
	}
}
*/

func NewOrderTime() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("OrderTime")
	res.Map = `from order in docs.Orders
select new {
  delay = order.shippedAt - ((DateTime?)order.orderedAt)
}`
	return res
}

func query_queryWithDuration(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	now := ravendb.Time(time.Now())

	index := NewOrderTime()
	err = store.ExecuteIndex(index)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		order1 := &Order{
			Company:   "hours",
			OrderedAt: addHours(now, -2),
			ShippedAt: now,
		}

		err = session.Store(order1)
		assert.NoError(t, err)

		order2 := &Order{
			Company:   "days",
			OrderedAt: addDays(now, -2),
			ShippedAt: now,
		}
		err = session.Store(order2)
		assert.NoError(t, err)

		order3 := &Order{
			Company:   "minutes",
			OrderedAt: addMinutes(now, -2),
			ShippedAt: now,
		}

		err = session.Store(order3)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	driver.waitForIndexing(store, "", 0)

	{
		session := openSessionMust(t, store)

		{
			var orders []*Order
			q := session.QueryInIndexOld(reflect.TypeOf(&Order{}), NewOrderTime())
			q = q.WhereLessThan("delay", time.Hour*3)
			err := q.ToList(&orders)
			assert.NoError(t, err)

			var delay []string
			for _, order := range orders {
				company := order.Company
				delay = append(delay, company)
			}
			sort.Strings(delay)
			stringArrayEq(delay, []string{"hours", "minutes"})
		}

		{
			var orders []*Order
			q := session.QueryInIndexOld(reflect.TypeOf(&Order{}), NewOrderTime())
			q = q.WhereGreaterThan("delay", time.Hour*3)
			err := q.ToList(&orders)
			assert.NoError(t, err)

			var delay2 []string
			for _, order := range orders {
				company := order.Company
				delay2 = append(delay2, company)
			}
			sort.Strings(delay2)
			stringArrayEq(delay2, []string{"days"})

		}

		session.Close()
	}
}

func query_queryFirst(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)
	{
		session := openSessionMust(t, store)

		var first *User
		err = session.Query().First(&first)
		assert.NoError(t, err)
		assert.NotNil(t, first)
		assert.Equal(t, first.ID, "users/1")

		single, err := session.QueryOld(reflect.TypeOf(&User{})).WhereEquals("name", "Tarzan").Single()
		assert.NoError(t, err)
		assert.NotNil(t, single)

		q := session.QueryOld(reflect.TypeOf(&User{}))
		_, err = q.Single()
		_ = err.(*ravendb.IllegalStateError)

		session.Close()
	}
}

func query_queryParameters(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)
	{
		session := openSessionMust(t, store)

		q := session.RawQuery("from Users where name = $name")
		q = q.AddParameter("name", "Tarzan")
		count, err := q.Count()
		assert.NoError(t, err)

		assert.Equal(t, count, 1)

		session.Close()
	}
}

func query_queryRandomOrder(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)
	{
		session := openSessionMust(t, store)
		{
			var res []*User
			q := session.Query().RandomOrdering()
			err := q.ToList(&res)
			assert.NoError(t, err)
			assert.Equal(t, len(res), 3)
		}

		{
			var res []*User
			q := session.Query().RandomOrderingWithSeed("123")
			err := q.ToList(&res)
			assert.NoError(t, err)
			assert.Equal(t, len(res), 3)
		}

		session.Close()
	}
}

func query_queryWhereExists(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)
	{
		session := openSessionMust(t, store)

		{
			var res []*User
			q := session.Query()
			q = q.WhereExists("name")
			err := q.ToList(&res)
			assert.NoError(t, err)
			assert.Equal(t, len(res), 3)
		}

		{
			var res []*User
			q := session.Query()
			q = q.WhereExists("name")
			q = q.AndAlso()
			q = q.Not()
			q = q.WhereExists("no_such_field")
			err := q.ToList(&res)
			assert.NoError(t, err)
			assert.Equal(t, len(res), 3)
		}

		session.Close()
	}
}

func query_queryWithBoost(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	query_addUsers(t, store, driver)
	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.Query()
		q = q.WhereEquals("name", "Tarzan")
		q = q.Boost(5)
		q = q.OrElse()
		q = q.WhereEquals("name", "John")
		q = q.Boost(2)
		q = q.OrderByScore()
		err := q.ToList(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 3)

		var names []string
		for _, user := range users {
			names = append(names, *user.Name)
		}
		assert.True(t, stringArrayContainsSequence(names, []string{"Tarzan", "John", "John"}))

		users = nil
		q = session.Query()
		q = q.WhereEquals("name", "Tarzan")
		q = q.Boost(2)
		q = q.OrElse()
		q = q.WhereEquals("name", "John")
		q = q.Boost(5)
		q = q.OrderByScore()
		err = q.ToList(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 3)

		names = nil
		for _, user := range users {
			names = append(names, *user.Name)
		}

		assert.True(t, stringArrayContainsSequence(names, []string{"John", "John", "Tarzan"}))

		session.Close()
	}
}

func makeUsersByNameIndex() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("UsersByName")
	res.Map = "from c in docs.Users select new " +
		" {" +
		"    c.name, " +
		"    count = 1" +
		"}"
	res.Reduce = "from result in results " +
		"group result by result.name " +
		"into g " +
		"select new " +
		"{ " +
		"  name = g.Key, " +
		"  count = g.Sum(x => x.count) " +
		"}"
	return res
}

func query_addUsers(t *testing.T, store *ravendb.IDocumentStore, driver *RavenTestDriver) {
	var err error

	{
		session := openSessionMust(t, store)
		user1 := &User{}
		user1.setName("John")
		user1.Age = 3

		user2 := &User{}
		user2.setName("John")
		user2.Age = 5

		user3 := &User{}
		user3.setName("Tarzan")
		user3.Age = 2

		err = session.StoreWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.StoreWithID(user2, "users/2")
		assert.NoError(t, err)
		err = session.StoreWithID(user3, "users/3")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	err = store.ExecuteIndex(makeUsersByNameIndex())
	assert.NoError(t, err)
	err = driver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)
}

func query_queryWithCustomize(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	err := store.ExecuteIndex(makeDogsIndex())
	assert.NoError(t, err)

	{
		newSession := openSessionMust(t, store)
		query_createDogs(t, newSession)
		err = newSession.SaveChanges()
		assert.NoError(t, err)

		newSession.Close()
	}

	{
		newSession := openSessionMust(t, store)

		var queryResult []*DogsIndex_Result
		q := newSession.Advanced().DocumentQueryAllOld(reflect.TypeOf(&DogsIndex_Result{}), "DogsIndex", "", false)
		q = q.WaitForNonStaleResults(0)
		q = q.OrderByWithOrdering("name", ravendb.OrderingType_ALPHA_NUMERIC)
		q = q.WhereGreaterThan("age", 2)
		err := q.ToList(&queryResult)
		assert.NoError(t, err)

		assert.Equal(t, len(queryResult), 4)

		r := queryResult[0]
		assert.Equal(t, r.Name, "Brian")

		r = queryResult[1]
		assert.Equal(t, r.Name, "Django")

		r = queryResult[2]
		assert.Equal(t, r.Name, "Lassie")

		r = queryResult[3]
		assert.Equal(t, r.Name, "Snoopy")

		newSession.Close()
	}
}

func query_createDogs(t *testing.T, newSession *ravendb.DocumentSession) {
	var err error

	dog1 := NewDog()
	dog1.Name = "Snoopy"
	dog1.Breed = "Beagle"
	dog1.Color = "White"
	dog1.Age = 6
	dog1.IsVaccinated = true

	err = newSession.StoreWithID(dog1, "docs/1")
	assert.NoError(t, err)

	dog2 := NewDog()
	dog2.Name = "Brian"
	dog2.Breed = "Labrador"
	dog2.Color = "White"
	dog2.Age = 12
	dog2.IsVaccinated = false

	err = newSession.StoreWithID(dog2, "docs/2")
	assert.NoError(t, err)

	dog3 := NewDog()
	dog3.Name = "Django"
	dog3.Breed = "Jack Russel"
	dog3.Color = "Black"
	dog3.Age = 3
	dog3.IsVaccinated = true

	err = newSession.StoreWithID(dog3, "docs/3")
	assert.NoError(t, err)

	dog4 := NewDog()
	dog4.Name = "Beethoven"
	dog4.Breed = "St. Bernard"
	dog4.Color = "Brown"
	dog4.Age = 1
	dog4.IsVaccinated = false

	err = newSession.StoreWithID(dog4, "docs/4")
	assert.NoError(t, err)

	dog5 := NewDog()
	dog5.Name = "Scooby Doo"
	dog5.Breed = "Great Dane"
	dog5.Color = "Brown"
	dog5.Age = 0
	dog5.IsVaccinated = false

	err = newSession.StoreWithID(dog5, "docs/5")
	assert.NoError(t, err)

	dog6 := NewDog()
	dog6.Name = "Old Yeller"
	dog6.Breed = "Black Mouth Cur"
	dog6.Color = "White"
	dog6.Age = 2
	dog6.IsVaccinated = true

	err = newSession.StoreWithID(dog6, "docs/6")
	assert.NoError(t, err)

	dog7 := NewDog()
	dog7.Name = "Benji"
	dog7.Breed = "Mixed"
	dog7.Color = "White"
	dog7.Age = 0
	dog7.IsVaccinated = false

	err = newSession.StoreWithID(dog7, "docs/7")
	assert.NoError(t, err)

	dog8 := NewDog()
	dog8.Name = "Lassie"
	dog8.Breed = "Collie"
	dog8.Color = "Brown"
	dog8.Age = 6
	dog8.IsVaccinated = true

	err = newSession.StoreWithID(dog8, "docs/8")
	assert.NoError(t, err)
}

type Dog struct {
	ID           string
	Name         string `json:"name"`
	Breed        string `json:"breed"`
	Color        string `json:"color"`
	Age          int    `json:"age"`
	IsVaccinated bool   `json:"vaccinated"`
}

func NewDog() *Dog {
	return &Dog{}
}

type DogsIndex_Result struct {
	Name         string `json:"name"`
	Age          int    `json:"age"`
	IsVaccinated bool   `json:"vaccinated"`
}

func makeDogsIndex() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("DogsIndex")
	res.Map = "from dog in docs.dogs select new { dog.name, dog.age, dog.vaccinated }"
	return res
}

func query_queryLongRequest(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)

		longName := strings.Repeat("x", 2048)
		user := &User{}
		user.setName(longName)
		err = newSession.StoreWithID(user, "users/1")
		assert.NoError(t, err)

		err = newSession.SaveChanges()
		assert.NoError(t, err)

		var queryResult []*User
		q := newSession.Advanced().DocumentQueryAllOld(reflect.TypeOf(&User{}), "", "Users", false)
		q = q.WhereEquals("name", longName)
		err := q.ToList(&queryResult)
		assert.NoError(t, err)
		assert.Equal(t, len(queryResult), 1)

		newSession.Close()
	}
}

func query_queryByIndex(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	err = store.ExecuteIndex(makeDogsIndex())
	assert.NoError(t, err)

	{
		newSession := openSessionMust(t, store)
		query_createDogs(t, newSession)

		err = newSession.SaveChanges()
		assert.NoError(t, err)

		err = driver.waitForIndexing(store, store.GetDatabase(), 0)
		assert.NoError(t, err)

		newSession.Close()
	}

	{
		newSession := openSessionMust(t, store)

		var queryResult []*DogsIndex_Result
		q := newSession.Advanced().DocumentQueryAllOld(reflect.TypeOf(&DogsIndex_Result{}), "DogsIndex", "", false)
		q = q.WhereGreaterThan("age", 2)
		q = q.AndAlso()
		q = q.WhereEquals("vaccinated", false)
		err := q.ToList(&queryResult)
		assert.NoError(t, err)

		assert.Equal(t, len(queryResult), 1)
		r := queryResult[0]
		assert.Equal(t, r.Name, "Brian")

		var queryResult2 []*DogsIndex_Result
		q = newSession.Advanced().DocumentQueryAllOld(reflect.TypeOf(&DogsIndex_Result{}), "DogsIndex", "", false)
		q = q.WhereLessThanOrEqual("age", 2)
		q = q.AndAlso()
		q = q.WhereEquals("vaccinated", false)
		err = q.ToList(&queryResult2)
		assert.NoError(t, err)

		assert.Equal(t, len(queryResult2), 3)

		var names []string
		for _, dir := range queryResult2 {
			name := dir.Name
			names = append(names, name)
		}
		sort.Strings(names)

		assert.True(t, stringArrayContainsSequence(names, []string{"Beethoven", "Benji", "Scooby Doo"}))
		newSession.Close()
	}
}

type ReduceResult struct {
	Count int    `json:"count"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
}

func TestQuery(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	query_queryWhereExists(t, driver)
	query_querySearchWithOr(t, driver)
	query_rawQuerySkipTake(t, driver)
	query_queryWithDuration(t, driver)
	query_queryWithWhereClause(t, driver)
	query_queryMapReduceIndex(t, driver)
	query_queryLazily(t, driver)
	query_queryLucene(t, driver)
	query_queryWithWhereGreaterThan(t, driver)
	query_querySimple(t, driver)
	query_queryWithSelect(t, driver)
	query_collectionsStats(t, driver)
	query_queryWithWhereBetween(t, driver)
	query_queryRandomOrder(t, driver)
	query_queryNoTracking(t, driver)
	query_queryLongRequest(t, driver)
	query_queryWithProjection2(t, driver)
	query_queryWhereNot(t, driver)
	query_querySkipTake(t, driver)
	query_queryWithProjection(t, driver)
	query_queryFirst(t, driver)
	query_querySingleProperty(t, driver)
	query_parametersInRawQuery(t, driver)
	query_queryWithWhereLessThan(t, driver)
	query_queryMapReduceWithCount(t, driver)
	query_queryWithWhereGreaterThanOrEqual(t, driver)
	query_queryWithCustomize(t, driver)
	query_queryWithBoost(t, driver)
	query_queryMapReduceWithSum(t, driver)
	query_queryWhereExact(t, driver)
	query_queryParameters(t, driver)
	query_queryByIndex(t, driver)
	query_queryWithWhereIn(t, driver)
	query_queryDistinct(t, driver)
	query_queryWithWhereLessThanOrEqual(t, driver)
}
