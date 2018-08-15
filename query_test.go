package ravendb

import (
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func query_querySimple(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user1 := NewUser()
		user1.setName("John")

		user2 := NewUser()
		user2.setName("Jane")

		user3 := NewUser()
		user3.setName("Tarzan")

		err = session.StoreWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.StoreWithID(user2, "users/2")
		assert.NoError(t, err)
		err = session.StoreWithID(user3, "users/3")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		q := session.Advanced().DocumentQueryAll(GetTypeOf(&User{}), "", "users", false)
		queryResult, err := q.toList()
		assert.NoError(t, err)
		assert.Equal(t, len(queryResult), 3)

		session.Close()
	}
}

// TODO: requires Lazy support
func query_queryLazily(t *testing.T) {}

func query_collectionsStats(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user1 := NewUser()
		user1.setName("John")

		user2 := NewUser()
		user2.setName("Jane")

		err = session.StoreWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.StoreWithID(user2, "users/2")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	op := NewGetCollectionStatisticsOperation()
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)
	stats := op.Command.Result
	assert.Equal(t, stats.getCountOfDocuments(), 2)
	coll := stats.getCollections()["Users"]
	assert.Equal(t, coll, 2)
}

func query_queryWithWhereClause(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user1 := NewUser()
		user1.setName("John")

		user2 := NewUser()
		user2.setName("Jane")

		user3 := NewUser()
		user3.setName("Tarzan")

		err = session.StoreWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.StoreWithID(user2, "users/2")
		assert.NoError(t, err)
		err = session.StoreWithID(user3, "users/3")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		q := session.QueryWithQuery(GetTypeOf(&User{}), Query_collection("users"))
		q = q.WhereStartsWith("name", "J")
		queryResult, err := q.toList()
		assert.NoError(t, err)

		q2 := session.QueryWithQuery(GetTypeOf(&User{}), Query_collection("users"))
		q2 = q2.WhereEquals("name", "Tarzan")
		queryResult2, err := q2.toList()
		assert.NoError(t, err)

		q3 := session.QueryWithQuery(GetTypeOf(&User{}), Query_collection("users"))
		q3 = q3.WhereEndsWith("name", "n")
		queryResult3, err := q3.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(queryResult), 2)
		assert.Equal(t, len(queryResult2), 1)
		assert.Equal(t, len(queryResult3), 2)

		session.Close()
	}
}

func query_queryMapReduceWithCount(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		q := session.Query(GetTypeOf(&User{}))
		q2 := q.GroupBy("name")
		q2 = q2.selectKey()
		q = q2.selectCount()
		q = q.OrderByDescending("count")
		q = q.OfType(GetTypeOf(&ReduceResult{}))
		results, err := q.toList()
		assert.NoError(t, err)

		{
			result := results[0].(*ReduceResult)
			assert.Equal(t, result.getCount(), 2)
			assert.Equal(t, result.getName(), "John")
		}

		{
			result := results[1].(*ReduceResult)
			assert.Equal(t, result.getCount(), 1)
			assert.Equal(t, result.getName(), "Tarzan")
		}

		session.Close()
	}
}

func query_queryMapReduceWithSum(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		q := session.Query(GetTypeOf(&User{}))
		q2 := q.GroupBy("name")
		q2 = q2.selectKey()
		q = q2.selectSum(NewGroupByFieldWithName("age"))
		q = q.OrderByDescending("age")
		q = q.OfType(GetTypeOf(&ReduceResult{}))
		results, err := q.toList()
		assert.NoError(t, err)

		{
			result := results[0].(*ReduceResult)
			assert.Equal(t, result.getAge(), 8)
			assert.Equal(t, result.getName(), "John")
		}

		{
			result := results[1].(*ReduceResult)
			assert.Equal(t, result.getAge(), 2)
			assert.Equal(t, result.getName(), "Tarzan")
		}

		session.Close()
	}
}

func query_queryMapReduceIndex(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		q := session.QueryWithQuery(GetTypeOf(&ReduceResult{}), Query_index("UsersByName"))
		q = q.OrderByDescending("count")
		results, err := q.toList()
		assert.NoError(t, err)

		{
			result := results[0].(*ReduceResult)
			assert.Equal(t, result.getCount(), 2)
			assert.Equal(t, result.getName(), "John")
		}

		{
			result := results[1].(*ReduceResult)
			assert.Equal(t, result.getCount(), 1)
			assert.Equal(t, result.getName(), "Tarzan")
		}

		session.Close()
	}
}

func query_querySingleProperty(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		q := session.Query(GetTypeOf(&User{}))
		q = q.AddOrderWithOrdering("age", true, OrderingType_LONG)
		q = q.SelectFields(GetTypeOf(int(0)), "age")
		ages, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(ages), 3)

		for _, n := range []int{5, 3, 2} {
			assert.True(t, interfaceArrayContains(ages, n))
		}

		session.Close()
	}
}

func query_queryWithSelect(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		q := session.Query(GetTypeOf(&User{}))
		q = q.SelectFields(GetTypeOf(&User{}), "age")
		usersAge, err := q.toList()
		assert.NoError(t, err)

		for _, u := range usersAge {
			user := u.(*User)

			assert.True(t, user.getAge() >= 0)
			assert.NotEmpty(t, user.getId())
		}

		session.Close()
	}
}

func query_queryWithWhereIn(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		q := session.Query(GetTypeOf(&User{}))
		q = q.WhereIn("name", []Object{"Tarzan", "no_such"})
		users, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)

		session.Close()
	}
}

func query_queryWithWhereBetween(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		q := session.Query(GetTypeOf(&User{}))
		q = q.WhereBetween("age", 4, 5)
		users, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)

		user := users[0].(*User)
		assert.Equal(t, *user.getName(), "John")

		session.Close()
	}
}

func query_queryWithWhereLessThan(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		q := session.Query(GetTypeOf(&User{}))
		q = q.WhereLessThan("age", 3)
		users, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)

		user := users[0].(*User)
		assert.Equal(t, *user.getName(), "Tarzan")

		session.Close()
	}
}

func query_queryWithWhereLessThanOrEqual(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		q := session.Query(GetTypeOf(&User{}))
		q = q.WhereLessThanOrEqual("age", 3)
		users, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(users), 2)

		session.Close()
	}
}

func query_queryWithWhereGreaterThan(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		q := session.Query(GetTypeOf(&User{}))
		q = q.WhereGreaterThan("age", 3)
		users, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)

		user := users[0].(*User)
		assert.Equal(t, *user.getName(), "John")

		session.Close()
	}
}

func query_queryWithWhereGreaterThanOrEqual(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		q := session.Query(GetTypeOf(&User{}))
		q = q.WhereGreaterThanOrEqual("age", 3)
		users, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(users), 2)

		session.Close()
	}
}

type UserProjection struct {
	ID   string
	Name string
}

func (p *UserProjection) getId() string {
	return p.ID
}

func (p *UserProjection) setId(id string) {
	p.ID = id
}

func (p *UserProjection) getName() string {
	return p.Name
}

func (p *UserProjection) setName(name string) {
	p.Name = name
}

func query_queryWithProjection(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		q := session.Query(GetTypeOf(&User{}))
		q = q.SelectFields(GetTypeOf(&UserProjection{}))
		projections, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(projections), 3)

		for _, p := range projections {
			projection := p.(*UserProjection)
			assert.NotEmpty(t, projection.getId())

			assert.NotEmpty(t, projection.getName())
		}

		session.Close()
	}
}

func query_queryWithProjection2(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		q := session.Query(GetTypeOf(&User{}))
		q = q.SelectFields(GetTypeOf(&UserProjection{}), "lastName")
		projections, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(projections), 3)

		for _, p := range projections {
			projection := p.(*UserProjection)
			assert.NotEmpty(t, projection.getId())

			assert.Empty(t, projection.getName()) // we didn't specify this field in mapping
		}

		session.Close()
	}
}

func query_queryDistinct(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		q := session.Query(GetTypeOf(&User{}))
		q = q.SelectFields(GetTypeOf(""), "name")
		q = q.Distinct()
		uniqueNames, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(uniqueNames), 2)
		assert.True(t, interfaceArrayContains(uniqueNames, "Tarzan"))
		assert.True(t, interfaceArrayContains(uniqueNames, "John"))

		session.Close()
	}
}

func query_querySearchWithOr(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		q := session.Query(GetTypeOf(&User{}))
		q = q.SearchWithOperator("name", "Tarzan John", SearchOperator_OR)
		uniqueNames, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(uniqueNames), 3)

		session.Close()
	}
}

func query_queryNoTracking(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		q := session.Query(GetTypeOf(&User{}))
		q = q.NoTracking()
		users, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(users), 3)

		for _, u := range users {
			user := u.(*User)
			isLoaded := session.IsLoaded(user.getId())
			assert.False(t, isLoaded)
		}

		session.Close()
	}
}

func query_querySkipTake(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		q := session.Query(GetTypeOf(&User{}))
		q = q.OrderBy("name")
		q = q.Skip(2)
		q = q.Take(1)
		users, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)

		user := users[0].(*User)
		assert.Equal(t, *user.getName(), "Tarzan")

		session.Close()
	}
}

func query_rawQuerySkipTake(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		q := session.RawQuery(GetTypeOf(&User{}), "from users")
		q = q.skip(2)
		q = q.take(1)
		users, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)
		user := users[0].(*User)
		assert.Equal(t, *user.getName(), "Tarzan")

		session.Close()
	}
}

func query_parametersInRawQuery(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		q := session.RawQuery(GetTypeOf(&User{}), "from users where age == $p0")
		q = q.addParameter("p0", 5)
		users, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)
		user := users[0].(*User)
		assert.Equal(t, *user.getName(), "John")

		session.Close()
	}
}

func query_queryLucene(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		q := session.Query(GetTypeOf(&User{}))
		q = q.WhereLucene("name", "Tarzan")
		users, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)

		for _, u := range users {
			user := u.(*User)
			assert.Equal(t, *user.getName(), "Tarzan")
		}

		session.Close()
	}
}

func query_queryWhereExact(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		{
			q := session.Query(GetTypeOf(&User{}))
			q = q.WhereEquals("name", "tarzan")
			users, err := q.toList()
			assert.NoError(t, err)

			assert.Equal(t, len(users), 1)
		}

		{
			q := session.Query(GetTypeOf(&User{}))
			q = q.WhereEqualsWithExact("name", "tarzan", true)
			users, err := q.toList()
			assert.NoError(t, err)

			assert.Equal(t, len(users), 0) // we queried for tarzan with exact
		}

		{
			q := session.Query(GetTypeOf(&User{}))
			q = q.WhereEqualsWithExact("name", "Tarzan", true)
			users, err := q.toList()
			assert.NoError(t, err)

			assert.Equal(t, len(users), 1) // we queried for Tarzan with exact
		}

		session.Close()
	}
}

func query_queryWhereNot(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		{
			q := session.Query(GetTypeOf(&User{}))
			q = q.Not()
			q = q.WhereEquals("name", "tarzan")
			res, err := q.toList()

			assert.NoError(t, err)

			assert.Equal(t, len(res), 2)
		}

		{
			q := session.Query(GetTypeOf(&User{}))
			q = q.WhereNotEquals("name", "tarzan")
			res, err := q.toList()

			assert.NoError(t, err)

			assert.Equal(t, len(res), 2)
		}

		{
			q := session.Query(GetTypeOf(&User{}))
			q = q.WhereNotEqualsWithExact("name", "Tarzan", true)
			res, err := q.toList()

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

func NewOrderTime() *AbstractIndexCreationTask {
	res := NewAbstractIndexCreationTask("OrderTime")
	res.Map = `from order in docs.Orders
select new {
  delay = order.shippedAt - ((DateTime?)order.orderedAt)
}`
	return res
}

func query_queryWithDuration(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	now := time.Now()

	index := NewOrderTime()
	err = store.ExecuteIndex(index)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		order1 := NewOrder()
		order1.setCompany("hours")
		order1.setOrderedAt(DateUtils_addHours(now, -2))
		order1.setShippedAt(now)
		err = session.Store(order1)
		assert.NoError(t, err)

		order2 := NewOrder()
		order2.setCompany("days")
		order2.setOrderedAt(DateUtils_addDays(now, -2))
		order2.setShippedAt(now)
		err = session.Store(order2)
		assert.NoError(t, err)

		order3 := NewOrder()
		order3.setCompany("minutes")
		order3.setOrderedAt(DateUtils_addMinutes(now, -2))
		order3.setShippedAt(now)
		err = session.Store(order3)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	gRavenTestDriver.waitForIndexing(store, "", 0)

	{
		session := openSessionMust(t, store)

		{
			q := session.QueryInIndex(GetTypeOf(&Order{}), NewOrderTime())
			q = q.WhereLessThan("delay", time.Hour*3)
			orders, err := q.toList()
			assert.NoError(t, err)

			var delay []string
			for _, o := range orders {
				order := o.(*Order)
				company := order.getCompany()
				delay = append(delay, company)
			}
			sort.Strings(delay)
			stringArrayEq(delay, []string{"hours", "minutes"})
		}

		{
			q := session.QueryInIndex(GetTypeOf(&Order{}), NewOrderTime())
			q = q.WhereGreaterThan("delay", time.Hour*3)
			orders, err := q.toList()
			assert.NoError(t, err)

			var delay2 []string
			for _, o := range orders {
				order := o.(*Order)
				company := order.getCompany()
				delay2 = append(delay2, company)
			}
			sort.Strings(delay2)
			stringArrayEq(delay2, []string{"days"})

		}

		session.Close()
	}
}

func query_queryFirst(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)
	{
		session := openSessionMust(t, store)

		first, err := session.Query(GetTypeOf(&User{})).first()
		assert.NoError(t, err)
		assert.NotNil(t, first)

		single, err := session.Query(GetTypeOf(&User{})).WhereEquals("name", "Tarzan").single()
		assert.NoError(t, err)
		assert.NotNil(t, single)

		_, err = session.Query(GetTypeOf(&User{})).single()
		_ = err.(*IllegalStateException)

		session.Close()
	}
}

func query_queryParameters(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)
	{
		session := openSessionMust(t, store)

		q := session.RawQuery(GetTypeOf(&User{}), "from Users where name = $name")
		q = q.addParameter("name", "Tarzan")
		count, err := q.count()
		assert.NoError(t, err)

		assert.Equal(t, count, 1)

		session.Close()
	}
}

func query_queryRandomOrder(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)
	{
		session := openSessionMust(t, store)
		{
			q := session.Query(GetTypeOf(&User{})).RandomOrdering()
			res, err := q.toList()
			assert.NoError(t, err)
			assert.Equal(t, len(res), 3)
		}

		{
			q := session.Query(GetTypeOf(&User{})).RandomOrderingWithSeed("123")
			res, err := q.toList()
			assert.NoError(t, err)
			assert.Equal(t, len(res), 3)
		}

		session.Close()
	}
}

func query_queryWhereExists(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)
	{
		session := openSessionMust(t, store)

		{
			q := session.Query(GetTypeOf(&User{}))
			q = q.WhereExists("name")
			res, err := q.toList()
			assert.NoError(t, err)
			assert.Equal(t, len(res), 3)
		}

		{
			q := session.Query(GetTypeOf(&User{}))
			q = q.WhereExists("name")
			q = q.AndAlso()
			q = q.Not()
			q = q.WhereExists("no_such_field")
			res, err := q.toList()
			assert.NoError(t, err)
			assert.Equal(t, len(res), 3)
		}

		session.Close()
	}
}

func query_queryWithBoost(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)
	{
		session := openSessionMust(t, store)

		q := session.Query(GetTypeOf(&User{}))
		q = q.WhereEquals("name", "Tarzan")
		q = q.Boost(5)
		q = q.OrElse()
		q = q.WhereEquals("name", "John")
		q = q.Boost(2)
		q = q.OrderByScore()
		users, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(users), 3)

		var names []string
		for _, u := range users {
			user := u.(*User)
			names = append(names, *user.getName())
		}
		assert.True(t, stringArrayContainsSequence(names, []string{"Tarzan", "John", "John"}))

		q = session.Query(GetTypeOf(&User{}))
		q = q.WhereEquals("name", "Tarzan")
		q = q.Boost(2)
		q = q.OrElse()
		q = q.WhereEquals("name", "John")
		q = q.Boost(5)
		q = q.OrderByScore()
		users, err = q.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(users), 3)

		names = nil
		for _, u := range users {
			user := u.(*User)
			names = append(names, *user.getName())
		}

		assert.True(t, stringArrayContainsSequence(names, []string{"John", "John", "Tarzan"}))

		session.Close()
	}
}

func makeUsersByNameIndex() *AbstractIndexCreationTask {
	res := NewAbstractIndexCreationTask("UsersByName")
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

func query_addUsers(t *testing.T, store *IDocumentStore) {
	var err error

	{
		session := openSessionMust(t, store)
		user1 := NewUser()
		user1.setName("John")
		user1.setAge(3)

		user2 := NewUser()
		user2.setName("John")
		user2.setAge(5)

		user3 := NewUser()
		user3.setName("Tarzan")
		user3.setAge(2)

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
	err = gRavenTestDriver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)
}

func query_queryWithCustomize(t *testing.T) {
	store := getDocumentStoreMust(t)
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

		q := newSession.Advanced().DocumentQueryAll(GetTypeOf(&DogsIndex_Result{}), "DogsIndex", "", false)
		q = q.WaitForNonStaleResults(0)
		q = q.OrderByWithOrdering("name", OrderingType_ALPHA_NUMERIC)
		q = q.WhereGreaterThan("age", 2)
		queryResult, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(queryResult), 4)

		r := queryResult[0].(*DogsIndex_Result)
		assert.Equal(t, r.getName(), "Brian")

		r = queryResult[1].(*DogsIndex_Result)
		assert.Equal(t, r.getName(), "Django")

		r = queryResult[2].(*DogsIndex_Result)
		assert.Equal(t, r.getName(), "Lassie")

		r = queryResult[3].(*DogsIndex_Result)
		assert.Equal(t, r.getName(), "Snoopy")

		newSession.Close()
	}
}

func query_createDogs(t *testing.T, newSession *DocumentSession) {
	var err error

	dog1 := NewDog()
	dog1.setName("Snoopy")
	dog1.setBreed("Beagle")
	dog1.setColor("White")
	dog1.setAge(6)
	dog1.setVaccinated(true)

	err = newSession.StoreWithID(dog1, "docs/1")
	assert.NoError(t, err)

	dog2 := NewDog()
	dog2.setName("Brian")
	dog2.setBreed("Labrador")
	dog2.setColor("White")
	dog2.setAge(12)
	dog2.setVaccinated(false)

	err = newSession.StoreWithID(dog2, "docs/2")
	assert.NoError(t, err)

	dog3 := NewDog()
	dog3.setName("Django")
	dog3.setBreed("Jack Russel")
	dog3.setColor("Black")
	dog3.setAge(3)
	dog3.setVaccinated(true)

	err = newSession.StoreWithID(dog3, "docs/3")
	assert.NoError(t, err)

	dog4 := NewDog()
	dog4.setName("Beethoven")
	dog4.setBreed("St. Bernard")
	dog4.setColor("Brown")
	dog4.setAge(1)
	dog4.setVaccinated(false)

	err = newSession.StoreWithID(dog4, "docs/4")
	assert.NoError(t, err)

	dog5 := NewDog()
	dog5.setName("Scooby Doo")
	dog5.setBreed("Great Dane")
	dog5.setColor("Brown")
	dog5.setAge(0)
	dog5.setVaccinated(false)

	err = newSession.StoreWithID(dog5, "docs/5")
	assert.NoError(t, err)

	dog6 := NewDog()
	dog6.setName("Old Yeller")
	dog6.setBreed("Black Mouth Cur")
	dog6.setColor("White")
	dog6.setAge(2)
	dog6.setVaccinated(true)

	err = newSession.StoreWithID(dog6, "docs/6")
	assert.NoError(t, err)

	dog7 := NewDog()
	dog7.setName("Benji")
	dog7.setBreed("Mixed")
	dog7.setColor("White")
	dog7.setAge(0)
	dog7.setVaccinated(false)

	err = newSession.StoreWithID(dog7, "docs/7")
	assert.NoError(t, err)

	dog8 := NewDog()
	dog8.setName("Lassie")
	dog8.setBreed("Collie")
	dog8.setColor("Brown")
	dog8.setAge(6)
	dog8.setVaccinated(true)

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

func (d *Dog) getId() string {
	return d.ID
}

func (d *Dog) setId(id string) {
	d.ID = id
}

func (d *Dog) getName() string {
	return d.Name
}

func (d *Dog) setName(name string) {
	d.Name = name
}

func (d *Dog) getBreed() string {
	return d.Breed
}

func (d *Dog) setBreed(breed string) {
	d.Breed = breed
}

func (d *Dog) getColor() string {
	return d.Color
}

func (d *Dog) setColor(color string) {
	d.Color = color
}

func (d *Dog) getAge() int {
	return d.Age
}

func (d *Dog) setAge(age int) {
	d.Age = age
}

func (d *Dog) isVaccinated() bool {
	return d.IsVaccinated
}

func (d *Dog) setVaccinated(vaccinated bool) {
	d.IsVaccinated = vaccinated
}

type DogsIndex_Result struct {
	Name         string `json:"name"`
	Age          int    `json:"age"`
	IsVaccinated bool   `json:"vaccinated"`
}

func (r *DogsIndex_Result) getName() string {
	return r.Name
}

func (r *DogsIndex_Result) setName(name string) {
	r.Name = name
}

func (r *DogsIndex_Result) getAge() int {
	return r.Age
}

func (r *DogsIndex_Result) setAge(age int) {
	r.Age = age
}

func (r *DogsIndex_Result) isVaccinated() bool {
	return r.IsVaccinated
}

func (r *DogsIndex_Result) setVaccinated(vaccinated bool) {
	r.IsVaccinated = vaccinated
}

func makeDogsIndex() *AbstractIndexCreationTask {
	res := NewAbstractIndexCreationTask("DogsIndex")
	res.Map = "from dog in docs.dogs select new { dog.name, dog.age, dog.vaccinated }"
	return res
}

func query_queryLongRequest(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)

		longName := strings.Repeat("x", 2048)
		user := NewUser()
		user.setName(longName)
		err = newSession.StoreWithID(user, "users/1")
		assert.NoError(t, err)

		err = newSession.SaveChanges()
		assert.NoError(t, err)

		q := newSession.Advanced().DocumentQueryAll(GetTypeOf(&User{}), "", "Users", false)
		q = q.WhereEquals("name", longName)
		queryResult, err := q.toList()
		assert.NoError(t, err)
		assert.Equal(t, len(queryResult), 1)

		newSession.Close()
	}
}

func query_queryByIndex(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	err = store.ExecuteIndex(makeDogsIndex())
	assert.NoError(t, err)

	{
		newSession := openSessionMust(t, store)
		query_createDogs(t, newSession)

		err = newSession.SaveChanges()
		assert.NoError(t, err)

		err = gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
		assert.NoError(t, err)

		newSession.Close()
	}

	{
		newSession := openSessionMust(t, store)

		q := newSession.Advanced().DocumentQueryAll(GetTypeOf(&DogsIndex_Result{}), "DogsIndex", "", false)
		q = q.WhereGreaterThan("age", 2)
		q = q.AndAlso()
		q = q.WhereEquals("vaccinated", false)
		queryResult, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(queryResult), 1)
		r := queryResult[0].(*DogsIndex_Result)
		assert.Equal(t, r.getName(), "Brian")

		q = newSession.Advanced().DocumentQueryAll(GetTypeOf(&DogsIndex_Result{}), "DogsIndex", "", false)
		q = q.WhereLessThanOrEqual("age", 2)
		q = q.AndAlso()
		q = q.WhereEquals("vaccinated", false)
		queryResult2, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(queryResult2), 3)

		var names []string
		for _, r := range queryResult2 {
			dir := r.(*DogsIndex_Result)
			name := dir.getName()
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

func (r *ReduceResult) getAge() int {
	return r.Age
}

func (r *ReduceResult) setAge(age int) {
	r.Age = age
}

func (r *ReduceResult) getCount() int {
	return r.Count
}

func (r *ReduceResult) setCount(count int) {
	r.Count = count
}

func (r *ReduceResult) getName() string {
	return r.Name
}

func (r *ReduceResult) setName(name string) {
	r.Name = name
}

func TestQuery(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	query_queryWhereExists(t)
	query_querySearchWithOr(t)
	//TODO: this test is flaky
	if gEnableFlakyTests {
		query_rawQuerySkipTake(t)
	}
	query_queryWithDuration(t)
	query_queryWithWhereClause(t)
	query_queryMapReduceIndex(t)
	query_queryLazily(t)
	query_queryLucene(t)
	query_queryWithWhereGreaterThan(t)
	query_querySimple(t)
	query_queryWithSelect(t)
	query_collectionsStats(t)
	query_queryWithWhereBetween(t)
	query_queryRandomOrder(t)
	query_queryNoTracking(t)
	query_queryLongRequest(t)
	query_queryWithProjection2(t)
	query_queryWhereNot(t)
	query_querySkipTake(t)
	query_queryWithProjection(t)
	query_queryFirst(t)
	query_querySingleProperty(t)
	//TODO: this test is flaky
	if gEnableFlakyTests {
		query_parametersInRawQuery(t)
	}
	query_queryWithWhereLessThan(t)
	query_queryMapReduceWithCount(t)
	query_queryWithWhereGreaterThanOrEqual(t)
	query_queryWithCustomize(t)
	query_queryWithBoost(t)
	query_queryMapReduceWithSum(t)
	query_queryWhereExact(t)
	query_queryParameters(t)
	query_queryByIndex(t)
	query_queryWithWhereIn(t)
	query_queryDistinct(t)
	query_queryWithWhereLessThanOrEqual(t)
}
