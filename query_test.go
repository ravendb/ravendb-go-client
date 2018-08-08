package ravendb

import (
	"fmt"
	"runtime/debug"
	"sort"
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

		err = session.StoreEntityWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.StoreEntityWithID(user2, "users/2")
		assert.NoError(t, err)
		err = session.StoreEntityWithID(user3, "users/3")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		q := session.advanced().documentQueryAll(getTypeOf(&User{}), "", "users", false)
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

		err = session.StoreEntityWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.StoreEntityWithID(user2, "users/2")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	op := NewGetCollectionStatisticsOperation()
	err = store.maintenance().send(op)
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

		err = session.StoreEntityWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.StoreEntityWithID(user2, "users/2")
		assert.NoError(t, err)
		err = session.StoreEntityWithID(user3, "users/3")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		q := session.queryWithQuery(getTypeOf(&User{}), Query_collection("users"))
		q = q.whereStartsWith("name", "J")
		queryResult, err := q.toList()
		assert.NoError(t, err)

		q2 := session.queryWithQuery(getTypeOf(&User{}), Query_collection("users"))
		q2 = q2.whereEquals("name", "Tarzan")
		queryResult2, err := q2.toList()
		assert.NoError(t, err)

		q3 := session.queryWithQuery(getTypeOf(&User{}), Query_collection("users"))
		q3 = q3.whereEndsWith("name", "n")
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

		q := session.query(getTypeOf(&User{}))
		q2 := q.groupBy("name")
		q2 = q2.selectKey()
		q = q2.selectCount()
		q = q.orderByDescending("count")
		q = q.ofType(getTypeOf(&ReduceResult{}))
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

		q := session.query(getTypeOf(&User{}))
		q2 := q.groupBy("name")
		q2 = q2.selectKey()
		q = q2.selectSum(NewGroupByFieldWithName("age"))
		q = q.orderByDescending("age")
		q = q.ofType(getTypeOf(&ReduceResult{}))
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

		q := session.queryWithQuery(getTypeOf(&ReduceResult{}), Query_index("UsersByName"))
		q = q.orderByDescending("count")
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

		q := session.query(getTypeOf(&User{}))
		q = q.addOrderWithOrdering("age", true, OrderingType_LONG)
		q = q.selectFields(getTypeOf(int(0)), "age")
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

		q := session.query(getTypeOf(&User{}))
		q = q.selectFields(getTypeOf(&User{}), "age")
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

		q := session.query(getTypeOf(&User{}))
		q = q.whereIn("name", []Object{"Tarzan", "no_such"})
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

		q := session.query(getTypeOf(&User{}))
		q = q.whereBetween("age", 4, 5)
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

		q := session.query(getTypeOf(&User{}))
		q = q.whereLessThan("age", 3)
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

		q := session.query(getTypeOf(&User{}))
		q = q.whereLessThanOrEqual("age", 3)
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

		q := session.query(getTypeOf(&User{}))
		q = q.whereGreaterThan("age", 3)
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

		q := session.query(getTypeOf(&User{}))
		q = q.whereGreaterThanOrEqual("age", 3)
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

		q := session.query(getTypeOf(&User{}))
		q = q.selectFields(getTypeOf(&UserProjection{}))
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

		q := session.query(getTypeOf(&User{}))
		q = q.selectFields(getTypeOf(&UserProjection{}), "lastName")
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

		q := session.query(getTypeOf(&User{}))
		q = q.selectFields(getTypeOf(""), "name")
		q = q.distinct()
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

		q := session.query(getTypeOf(&User{}))
		q = q.searchWithOperator("name", "Tarzan John", SearchOperator_OR)
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

		q := session.query(getTypeOf(&User{}))
		q = q.noTracking()
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

		q := session.query(getTypeOf(&User{}))
		q = q.orderBy("name")
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

func query_rawQuerySkipTake(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	query_addUsers(t, store)

	{
		session := openSessionMust(t, store)

		q := session.rawQuery(getTypeOf(&User{}), "from users")
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

		q := session.rawQuery(getTypeOf(&User{}), "from users where age == $p0")
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

		q := session.query(getTypeOf(&User{}))
		q = q.whereLucene("name", "Tarzan")
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
			q := session.query(getTypeOf(&User{}))
			q = q.whereEquals("name", "tarzan")
			users, err := q.toList()
			assert.NoError(t, err)

			assert.Equal(t, len(users), 1)
		}

		{
			q := session.query(getTypeOf(&User{}))
			q = q.whereEqualsWithExact("name", "tarzan", true)
			users, err := q.toList()
			assert.NoError(t, err)

			assert.Equal(t, len(users), 0) // we queried for tarzan with exact
		}

		{
			q := session.query(getTypeOf(&User{}))
			q = q.whereEqualsWithExact("name", "Tarzan", true)
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
			q := session.query(getTypeOf(&User{}))
			q = q.not()
			q = q.whereEquals("name", "tarzan")
			res, err := q.toList()

			assert.NoError(t, err)

			assert.Equal(t, len(res), 2)
		}

		{
			q := session.query(getTypeOf(&User{}))
			q = q.whereNotEquals("name", "tarzan")
			res, err := q.toList()

			assert.NoError(t, err)

			assert.Equal(t, len(res), 2)
		}

		{
			q := session.query(getTypeOf(&User{}))
			q = q.whereNotEqualsWithExact("name", "Tarzan", true)
			res, err := q.toList()

			assert.NoError(t, err)

			assert.Equal(t, len(res), 2)
		}

		session.Close()
	}
}

/*
TODO: is this used?
public static class Result {
	private long delay;

	public long getDelay() {
		return delay;
	}

	public void setDelay(long delay) {
		this.delay = delay;
	}
}
*/

func NewOrderTime() *AbstractIndexCreationTask {
	res := NewAbstractIndexCreationTask("OrderTime")
	res.smap = `from order in docs.Orders
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
	err = store.executeIndex(index)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		order1 := NewOrder()
		order1.setCompany("hours")
		order1.setOrderedAt(DateUtils_addHours(now, -2))
		order1.setShippedAt(now)
		err = session.StoreEntity(order1)
		assert.NoError(t, err)

		order2 := NewOrder()
		order2.setCompany("days")
		order2.setOrderedAt(DateUtils_addDays(now, -2))
		order2.setShippedAt(now)
		err = session.StoreEntity(order2)
		assert.NoError(t, err)

		order3 := NewOrder()
		order3.setCompany("minutes")
		order3.setOrderedAt(DateUtils_addMinutes(now, -2))
		order3.setShippedAt(now)
		err = session.StoreEntity(order3)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	gRavenTestDriver.waitForIndexing(store, "", 0)

	{
		session := openSessionMust(t, store)

		{
			q := session.queryInIndex(getTypeOf(&Order{}), NewOrderTime())
			q = q.whereLessThan("delay", time.Hour*3)
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
			q := session.queryInIndex(getTypeOf(&Order{}), NewOrderTime())
			q = q.whereGreaterThan("delay", time.Hour*3)
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

		first, err := session.query(getTypeOf(&User{})).first()
		assert.NoError(t, err)
		assert.NotNil(t, first)

		single, err := session.query(getTypeOf(&User{})).whereEquals("name", "Tarzan").single()
		assert.NoError(t, err)
		assert.NotNil(t, single)

		_, err = session.query(getTypeOf(&User{})).single()
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

		q := session.rawQuery(getTypeOf(&User{}), "from Users where name = $name")
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
			q := session.query(getTypeOf(&User{})).randomOrdering()
			res, err := q.toList()
			assert.NoError(t, err)
			assert.Equal(t, len(res), 3)
		}

		{
			q := session.query(getTypeOf(&User{})).randomOrderingWithSeed("123")
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
			q := session.query(getTypeOf(&User{}))
			q = q.whereExists("name")
			res, err := q.toList()
			assert.NoError(t, err)
			assert.Equal(t, len(res), 3)
		}

		{
			q := session.query(getTypeOf(&User{}))
			q = q.whereExists("name")
			q = q.andAlso()
			q = q.not()
			q = q.whereExists("no_such_field")
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

		q := session.query(getTypeOf(&User{}))
		q = q.whereEquals("name", "Tarzan")
		q = q.boost(5)
		q = q.orElse()
		q = q.whereEquals("name", "John")
		q = q.boost(2)
		q = q.orderByScore()
		users, err := q.toList()
		assert.NoError(t, err)

		assert.Equal(t, len(users), 3)

		var names []string
		for _, u := range users {
			user := u.(*User)
			names = append(names, *user.getName())
		}
		assert.True(t, stringArrayContainsSequence(names, []string{"Tarzan", "John", "John"}))

		q = session.query(getTypeOf(&User{}))
		q = q.whereEquals("name", "Tarzan")
		q = q.boost(2)
		q = q.orElse()
		q = q.whereEquals("name", "John")
		q = q.boost(5)
		q = q.orderByScore()
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

func makeUsersByName() *AbstractIndexCreationTask {
	res := NewAbstractIndexCreationTask("UsersByName")
	res.smap = "from c in docs.Users select new " +
		" {" +
		"    c.name, " +
		"    count = 1" +
		"}"
	res.reduce = "from result in results " +
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

		err = session.StoreEntityWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.StoreEntityWithID(user2, "users/2")
		assert.NoError(t, err)
		err = session.StoreEntityWithID(user3, "users/3")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	err = store.executeIndex(makeUsersByName())
	assert.NoError(t, err)
	err = gRavenTestDriver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)
}

func query_queryWithCustomize(t *testing.T) {}
func query_queryLongRequest(t *testing.T)   {}
func query_queryByIndex(t *testing.T)       {}

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

	if false {
		oldDumpFailedHTTP := dumpFailedHTTP
		dumpFailedHTTP = true
		defer func() {
			dumpFailedHTTP = oldDumpFailedHTTP
		}()
	}

	destroyDriver := createTestDriver(t)
	defer func() {
		r := recover()
		destroyDriver()
		if r != nil {
			fmt.Printf("Panic: '%v'\n", r)
			debug.PrintStack()
			t.Fail()
		}
	}()

	// matches order of Java tests
	query_queryWhereExists(t)
	query_querySearchWithOr(t)
	//TODO: this test is flaky
	//query_rawQuerySkipTake(t)
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
	//query_parametersInRawQuery(t)
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
