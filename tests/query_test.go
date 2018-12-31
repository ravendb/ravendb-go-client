package tests

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func queryQuerySimple(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
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

		q := session.Advanced().DocumentQueryAll("", "users", false)
		var queryResult []*User
		err := q.GetResults(&queryResult)
		assert.NoError(t, err)
		assert.Equal(t, len(queryResult), 3)

		session.Close()
	}
}

func queryQueryLazily(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
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

func queryCollectionsStats(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
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

func queryQueryWithWhereClause(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
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
		queryUsers := &ravendb.Query{
			Collection: "users",
		}
		q := session.QueryWithQuery(queryUsers)
		q = q.WhereStartsWith("name", "J")
		err := q.GetResults(&queryResult)
		assert.NoError(t, err)

		var queryResult2 []*User
		q2 := session.QueryWithQuery(queryUsers)
		q2 = q2.WhereEquals("name", "Tarzan")
		err = q2.GetResults(&queryResult2)
		assert.NoError(t, err)

		var queryResult3 []*User
		q3 := session.QueryWithQuery(queryUsers)
		q3 = q3.WhereEndsWith("name", "n")
		err = q3.GetResults(&queryResult3)
		assert.NoError(t, err)

		assert.Equal(t, len(queryResult), 2)
		assert.Equal(t, len(queryResult2), 1)
		assert.Equal(t, len(queryResult3), 2)

		session.Close()
	}
}

func queryQueryMapReduceWithCount(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var results []*ReduceResult
		q := session.QueryType(reflect.TypeOf(&User{}))
		q2 := q.GroupBy("name")
		q2 = q2.SelectKey()
		q = q2.SelectCount()
		q = q.OrderByDescending("count")
		err := q.GetResults(&results)
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

func queryQueryMapReduceWithSum(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)

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
		err := q.GetResults(&results)
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

func queryQueryMapReduceIndex(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var results []*ReduceResult
		queryIndex := &ravendb.Query{
			IndexName: "UsersByName",
		}
		q := session.QueryWithQuery(queryIndex)
		q = q.OrderByDescending("count")
		err := q.GetResults(&results)
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

func queryQuerySingleProperty(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		q := session.QueryType(reflect.TypeOf(&User{}))
		q = q.AddOrderWithOrdering("age", true, ravendb.OrderingTypeLong)
		q = q.SelectFields("age")
		var ages []int
		err := q.GetResults(&ages)
		assert.NoError(t, err)

		assert.Equal(t, len(ages), 3)

		for i, n := range []int{5, 3, 2} {
			assert.Equal(t, ages[i], n)
		}

		session.Close()
	}
}

func queryQueryWithSelect(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		q := session.Query()
		q = q.SelectFields("age")
		var usersAge []*User
		err := q.GetResults(&usersAge)
		assert.NoError(t, err)

		for _, user := range usersAge {
			assert.True(t, user.Age >= 0)
			assert.NotEmpty(t, user.ID)
		}

		session.Close()
	}
}

func queryQueryWithWhereIn(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.Query()
		q = q.WhereIn("name", []interface{}{"Tarzan", "no_such"})
		err := q.GetResults(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)

		session.Close()
	}
}

func queryQueryWithWhereBetween(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.Query()
		q = q.WhereBetween("age", 4, 5)
		err := q.GetResults(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)

		user := users[0]
		assert.Equal(t, *user.Name, "John")

		session.Close()
	}
}

func queryQueryWithWhereLessThan(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.Query()
		q = q.WhereLessThan("age", 3)
		err := q.GetResults(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)

		user := users[0]
		assert.Equal(t, *user.Name, "Tarzan")

		session.Close()
	}
}

func queryQueryWithWhereLessThanOrEqual(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.Query()
		q = q.WhereLessThanOrEqual("age", 3)
		err := q.GetResults(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 2)

		session.Close()
	}
}

func queryQueryWithWhereGreaterThan(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.Query()
		q = q.WhereGreaterThan("age", 3)
		err := q.GetResults(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)

		user := users[0]
		assert.Equal(t, *user.Name, "John")

		session.Close()
	}
}

func queryQueryWithWhereGreaterThanOrEqual(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.Query()
		q = q.WhereGreaterThanOrEqual("age", 3)
		err := q.GetResults(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 2)

		session.Close()
	}
}

type UserProjection struct {
	ID string
	// Note: this annotation is important because UsersByName
	// index uses lowercase "name" property
	Name string `json:"name"`
}

func queryQueryWithProjection(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		q := session.QueryType(reflect.TypeOf(&User{}))
		fields := ravendb.FieldsFor(&UserProjection{})
		q = q.SelectFields(fields...)
		var projections []*UserProjection
		err := q.GetResults(&projections)
		assert.NoError(t, err)

		assert.Equal(t, len(projections), 3)

		expectedNames := []string{"John", "John", "Tarzan"}
		for i, projection := range projections {
			expectedID := fmt.Sprintf("users/%d", i+1)
			assert.Equal(t, projection.ID, expectedID)
			assert.Equal(t, projection.Name, expectedNames[i])
		}

		session.Close()
	}
}

func queryQueryWithProjection2(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		q := session.QueryType(reflect.TypeOf(&User{}))
		q = q.SelectFields("lastName")
		var projections []*UserProjection
		err := q.GetResults(&projections)
		assert.NoError(t, err)

		assert.Equal(t, len(projections), 3)

		for _, projection := range projections {
			assert.NotEmpty(t, projection.ID)

			assert.Empty(t, projection.Name) // we didn't specify this field in mapping
		}

		session.Close()
	}
}

func queryQueryDistinct(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		q := session.QueryType(reflect.TypeOf(&User{}))
		q = q.SelectFields("name")
		q = q.Distinct()
		var uniqueNames []string
		err := q.GetResults(&uniqueNames)
		assert.NoError(t, err)

		assert.Equal(t, len(uniqueNames), 2)
		// TODO: not sure if order guaranteed. maybe sort before compare?
		assert.Equal(t, uniqueNames[0], "John")
		assert.Equal(t, uniqueNames[1], "Tarzan")

		session.Close()
	}
}

func queryQuerySearchWithOr(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var uniqueNames []*User
		q := session.Query()
		q = q.SearchWithOperator("name", "Tarzan John", ravendb.SearchOperator_OR)
		err := q.GetResults(&uniqueNames)
		assert.NoError(t, err)

		assert.Equal(t, len(uniqueNames), 3)

		session.Close()
	}
}

func queryQueryNoTracking(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.Query()
		q = q.NoTracking()
		err := q.GetResults(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 3)

		for _, user := range users {
			isLoaded := session.IsLoaded(user.ID)
			assert.False(t, isLoaded)
		}

		session.Close()
	}
}

func queryQuerySkipTake(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.Query()
		q = q.OrderBy("name")
		q = q.Skip(2)
		q = q.Take(1)
		err := q.GetResults(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)

		user := users[0]
		assert.Equal(t, *user.Name, "Tarzan")

		session.Close()
	}
}

func queryRawQuerySkipTake(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.RawQuery("from users")
		q = q.Skip(2)
		q = q.Take(1)
		err = q.GetResults(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)
		user := users[0]
		assert.Equal(t, *user.Name, "Tarzan")

		session.Close()
	}
}

func queryParametersInRawQuery(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.RawQuery("from users where age == $p0")
		q = q.AddParameter("p0", 5)
		err = q.GetResults(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)
		user := users[0]
		assert.Equal(t, *user.Name, "John")

		session.Close()
	}
}

func queryQueryLucene(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		var users []*User
		q := session.Query()
		q = q.WhereLucene("name", "Tarzan")
		err := q.GetResults(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)

		for _, user := range users {
			assert.Equal(t, *user.Name, "Tarzan")
		}

		session.Close()
	}
}

func queryQueryWhereExact(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		{
			var users []*User
			q := session.Query()
			q = q.WhereEquals("name", "tarzan")
			err := q.GetResults(&users)
			assert.NoError(t, err)

			assert.Equal(t, len(users), 1)
		}

		{
			var users []*User
			q := session.Query()
			q = q.WhereEquals("name", "tarzan").Exact()
			err := q.GetResults(&users)
			assert.NoError(t, err)

			assert.Equal(t, len(users), 0) // we queried for tarzan with exact
		}

		{
			var users []*User
			q := session.Query()
			q = q.WhereEquals("name", "Tarzan").Exact()
			err := q.GetResults(&users)
			assert.NoError(t, err)

			assert.Equal(t, len(users), 1) // we queried for Tarzan with exact
		}

		session.Close()
	}
}

func queryQueryWhereNot(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)

	{
		session := openSessionMust(t, store)

		{
			var res []*User
			q := session.Query()
			q = q.Not()
			q = q.WhereEquals("name", "tarzan")
			err := q.GetResults(&res)

			assert.NoError(t, err)

			assert.Equal(t, len(res), 2)
		}

		{
			var res []*User
			q := session.Query()
			q = q.WhereNotEquals("name", "tarzan")
			err := q.GetResults(&res)

			assert.NoError(t, err)

			assert.Equal(t, len(res), 2)
		}

		{
			var res []*User
			q := session.Query()
			q = q.WhereNotEquals("name", "Tarzan").Exact()
			err := q.GetResults(&res)

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

func queryQueryWithDuration(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
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
			q := session.QueryInIndex(NewOrderTime())
			q = q.WhereLessThan("delay", time.Hour*3)
			err := q.GetResults(&orders)
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
			q := session.QueryInIndex(NewOrderTime())
			q = q.WhereGreaterThan("delay", time.Hour*3)
			err := q.GetResults(&orders)
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

func queryQueryFirst(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)
	{
		session := openSessionMust(t, store)

		{
			var first *User
			err = session.Query().First(&first)
			assert.NoError(t, err)
			assert.NotNil(t, first)
			assert.Equal(t, first.ID, "users/1")
		}

		{
			var single *User
			q := session.Query().WhereEquals("name", "Tarzan")
			err = q.Single(&single)
			assert.NoError(t, err)
			assert.NotNil(t, single)
			assert.Equal(t, *single.Name, "Tarzan")
		}

		{
			var single *User
			q := session.Query()
			err = q.Single(&single)
			assert.Nil(t, single)
			_ = err.(*ravendb.IllegalStateError)
		}

		session.Close()
	}
}

func queryQueryParameters(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)
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

func queryQueryRandomOrder(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)
	{
		session := openSessionMust(t, store)
		{
			var res []*User
			q := session.Query().RandomOrdering()
			err := q.GetResults(&res)
			assert.NoError(t, err)
			assert.Equal(t, len(res), 3)
		}

		{
			var res []*User
			q := session.Query().RandomOrderingWithSeed("123")
			err := q.GetResults(&res)
			assert.NoError(t, err)
			assert.Equal(t, len(res), 3)
		}

		session.Close()
	}
}

func queryQueryWhereExists(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)
	{
		session := openSessionMust(t, store)

		{
			var res []*User
			q := session.Query()
			q = q.WhereExists("name")
			err := q.GetResults(&res)
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
			err := q.GetResults(&res)
			assert.NoError(t, err)
			assert.Equal(t, len(res), 3)
		}

		session.Close()
	}
}

func queryQueryWithBoost(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	queryAddUsers(t, store, driver)
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
		err := q.GetResults(&users)
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
		err = q.GetResults(&users)
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

func queryAddUsers(t *testing.T, store *ravendb.IDocumentStore, driver *RavenTestDriver) {
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

func queryQueryWithCustomize(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	err := store.ExecuteIndex(makeDogsIndex())
	assert.NoError(t, err)

	{
		newSession := openSessionMust(t, store)
		queryCreateDogs(t, newSession)
		err = newSession.SaveChanges()
		assert.NoError(t, err)

		newSession.Close()
	}

	{
		newSession := openSessionMust(t, store)

		q := newSession.Advanced().DocumentQueryAll("DogsIndex", "", false)
		q = q.WaitForNonStaleResults(0)
		q = q.OrderByWithOrdering("name", ravendb.OrderingTypeAlphaNumeric)
		q = q.WhereGreaterThan("age", 2)
		var queryResult []*DogsIndex_Result
		err := q.GetResults(&queryResult)
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

func queryCreateDogs(t *testing.T, newSession *ravendb.DocumentSession) {
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

func queryQueryLongRequest(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
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

		q := newSession.Advanced().DocumentQueryAll("", "Users", false)
		q = q.WhereEquals("name", longName)
		var queryResult []*User
		err := q.GetResults(&queryResult)
		assert.NoError(t, err)
		assert.Equal(t, len(queryResult), 1)

		newSession.Close()
	}
}

func queryQueryByIndex(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	err = store.ExecuteIndex(makeDogsIndex())
	assert.NoError(t, err)

	{
		newSession := openSessionMust(t, store)
		queryCreateDogs(t, newSession)

		err = newSession.SaveChanges()
		assert.NoError(t, err)

		err = driver.waitForIndexing(store, store.GetDatabase(), 0)
		assert.NoError(t, err)

		newSession.Close()
	}

	{
		newSession := openSessionMust(t, store)

		q := newSession.Advanced().DocumentQueryAll("DogsIndex", "", false)
		q = q.WhereGreaterThan("age", 2)
		q = q.AndAlso()
		q = q.WhereEquals("vaccinated", false)
		var queryResult []*DogsIndex_Result
		err := q.GetResults(&queryResult)
		assert.NoError(t, err)

		assert.Equal(t, len(queryResult), 1)
		r := queryResult[0]
		assert.Equal(t, r.Name, "Brian")

		q = newSession.Advanced().DocumentQueryAll("DogsIndex", "", false)
		q = q.WhereLessThanOrEqual("age", 2)
		q = q.AndAlso()
		q = q.WhereEquals("vaccinated", false)
		var queryResult2 []*DogsIndex_Result
		err = q.GetResults(&queryResult2)
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
	queryQueryWhereExists(t, driver)
	queryQuerySearchWithOr(t, driver)
	queryRawQuerySkipTake(t, driver)
	queryQueryWithDuration(t, driver)
	queryQueryWithWhereClause(t, driver)
	queryQueryMapReduceIndex(t, driver)
	queryQueryLazily(t, driver)
	queryQueryLucene(t, driver)
	queryQueryWithWhereGreaterThan(t, driver)
	queryQuerySimple(t, driver)
	queryQueryWithSelect(t, driver)
	queryCollectionsStats(t, driver)
	queryQueryWithWhereBetween(t, driver)
	queryQueryRandomOrder(t, driver)
	queryQueryNoTracking(t, driver)
	queryQueryLongRequest(t, driver)
	queryQueryWithProjection2(t, driver)
	queryQueryWhereNot(t, driver)
	queryQuerySkipTake(t, driver)
	queryQueryWithProjection(t, driver)
	queryQueryFirst(t, driver)
	queryQuerySingleProperty(t, driver)
	queryParametersInRawQuery(t, driver)
	queryQueryWithWhereLessThan(t, driver)
	queryQueryMapReduceWithCount(t, driver)
	queryQueryWithWhereGreaterThanOrEqual(t, driver)
	queryQueryWithCustomize(t, driver)
	queryQueryWithBoost(t, driver)
	queryQueryMapReduceWithSum(t, driver)
	queryQueryWhereExact(t, driver)
	queryQueryParameters(t, driver)
	queryQueryByIndex(t, driver)
	queryQueryWithWhereIn(t, driver)
	queryQueryDistinct(t, driver)
	queryQueryWithWhereLessThanOrEqual(t, driver)
}
