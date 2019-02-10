package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func lazyCanLazilyLoadEntity(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		for i := 1; i <= 6; i++ {
			company := &Company{
				ID: fmt.Sprintf("companies/%d", i),
			}
			err = session.StoreWithID(company, fmt.Sprintf("companies/%d", i))
			assert.NoError(t, err)
		}

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)
		query := session.Advanced().Lazily()
		var order *Company
		lazyOrder, err := query.Load(&order, "companies/1", nil)
		assert.NoError(t, err)

		assert.False(t, lazyOrder.IsValueCreated())
		err = lazyOrder.GetValue()
		assert.NoError(t, err)
		assert.Equal(t, order.ID, "companies/1")

		orders := map[string]*Company{}
		lazyOrders, err := session.Advanced().Lazily().LoadMulti(orders, []string{"companies/1", "companies/2"}, nil)
		assert.NoError(t, err)
		assert.False(t, lazyOrders.IsValueCreated())

		err = lazyOrders.GetValue()
		assert.NoError(t, err)
		assert.Equal(t, len(orders), 2)

		company1 := orders["companies/1"]
		company2 := orders["companies/2"]

		assert.NotNil(t, company1)
		assert.NotNil(t, company2)

		assert.Equal(t, company1.ID, "companies/1")

		assert.Equal(t, company2.ID, "companies/2")

		lazyOrder, err = session.Advanced().Lazily().Load(&order, "companies/3", nil)
		assert.NoError(t, err)
		assert.False(t, lazyOrder.IsValueCreated())

		err = lazyOrder.GetValue()
		assert.NoError(t, err)
		assert.Equal(t, order.ID, "companies/3")

		missingItems := map[string]*Company{}
		load, err := session.Advanced().Lazily().LoadMulti(missingItems, []string{"no_such_1", "no_such_2"}, nil)
		assert.NoError(t, err)
		err = load.GetValue()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(missingItems))

		assert.Nil(t, missingItems["no_such_1"])
		assert.Nil(t, missingItems["no_such_2"])

		session.Close()
	}
}

func lazyCanExecuteAllPendingLazyOperations(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		for i := 1; i <= 2; i++ {
			company := &Company{
				ID: fmt.Sprintf("companies/%d", i),
			}
			err = session.StoreWithID(company, company.ID)
			assert.NoError(t, err)
		}

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)
		var company1Ref *Company
		var company2Ref *Company
		query := session.Advanced().Lazily()
		var c1, c2 *Company
		_, err = query.Load(&c1, "companies/1", func(v interface{}) {
			c := v.(*Company)
			company1Ref = c
		})
		assert.NoError(t, err)

		_, err = query.Load(&c2, "companies/2", func(v interface{}) {
			c := v.(*Company)
			company2Ref = c
		})
		assert.NoError(t, err)

		assert.Nil(t, company1Ref)
		assert.Nil(t, company2Ref)

		_, err = session.Advanced().Eagerly().ExecuteAllPendingLazyOperations()
		assert.NoError(t, err)
		assert.Equal(t, company1Ref.ID, "companies/1")
		assert.Equal(t, company2Ref.ID, "companies/2")

		assert.Equal(t, c1, company1Ref)
		assert.Equal(t, c2, company2Ref)

		session.Close()
	}
}

func lazyWithQueuedActionsLoad(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user := &User{}
		user.setLastName("Oren")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var userRef *User
		var user *User

		query := session.Advanced().Lazily()
		_, err = query.Load(&user, "users/1", func(v interface{}) {
			userRef = v.(*User)
		})
		assert.NoError(t, err)

		assert.Nil(t, userRef)

		_, err = session.Advanced().Eagerly().ExecuteAllPendingLazyOperations()
		assert.NoError(t, err)
		assert.Equal(t, *userRef.LastName, "Oren")
		assert.Equal(t, user, userRef)

		session.Close()
	}
}

func lazyCanUseCacheWhenLazyLoading(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user := &User{}
		user.setLastName("Oren")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)
		var user *User
		lazyUser, err := session.Advanced().Lazily().Load(&user, "users/1", nil)
		assert.NoError(t, err)
		assert.False(t, lazyUser.IsValueCreated())

		err = lazyUser.GetValue()
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, user.ID, "users/1")

		session.Close()
	}

	{
		session := openSessionMust(t, store)
		var user *User
		lazyUser, err := session.Advanced().Lazily().Load(&user, "users/1", nil)
		assert.NoError(t, err)
		assert.False(t, lazyUser.IsValueCreated())

		err = lazyUser.GetValue()
		assert.NoError(t, err)
		assert.Equal(t, *user.LastName, "Oren")

		session.Close()
	}
}

func lazDontLazyLoadAlreadyLoadedValues(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user := &User{}
		user.setLastName("Oren")
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)

		user2 := &User{}
		user2.setLastName("Marcin")
		err = session.StoreWithID(user2, "users/2")
		assert.NoError(t, err)

		user3 := &User{}
		user3.setLastName("John")
		err = session.StoreWithID(user3, "users/3")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		users := map[string]*User{}
		lazyLoad, err := session.Advanced().Lazily().LoadMulti(users, []string{"users/2", "users/3"}, nil)
		assert.NoError(t, err)

		users2 := map[string]*User{}
		_, err = session.Advanced().Lazily().LoadMulti(users2, []string{"users/1", "users/3"}, nil)
		assert.NoError(t, err)

		var u1, u2 *User
		err = session.Load(&u1, "users/2")
		assert.NoError(t, err)
		err = session.Load(&u2, "users/3")
		assert.NoError(t, err)

		_, err = session.Advanced().Eagerly().ExecuteAllPendingLazyOperations()
		assert.NoError(t, err)

		assert.True(t, session.Advanced().IsLoaded("users/1"))

		err = lazyLoad.GetValue()
		assert.NoError(t, err)
		assert.Equal(t, len(users), 2)

		oldRequestCount := session.Advanced().GetNumberOfRequests()

		lazyLoad, err = session.Advanced().Lazily().LoadMulti(users, []string{"users/3"}, nil)
		assert.NoError(t, err)
		_, err = session.Advanced().Eagerly().ExecuteAllPendingLazyOperations()
		assert.NoError(t, err)

		assert.Equal(t, session.Advanced().GetNumberOfRequests(), oldRequestCount)

		session.Close()
	}
}

func TestLazy(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	lazyCanExecuteAllPendingLazyOperations(t, driver)
	lazyCanLazilyLoadEntity(t, driver)
	lazyCanUseCacheWhenLazyLoading(t, driver)
	lazyWithQueuedActionsLoad(t, driver)

	// TODO: order not same as Java
	lazDontLazyLoadAlreadyLoadedValues(t, driver)
}
