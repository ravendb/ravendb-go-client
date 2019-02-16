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
		lazyOrder, err := query.Load("companies/1")
		assert.NoError(t, err)

		assert.False(t, lazyOrder.IsValueCreated())
		var order *Company
		err = lazyOrder.GetValue(&order)
		assert.NoError(t, err)
		assert.Equal(t, order.ID, "companies/1")

		lazyOrders, err := session.Advanced().Lazily().LoadMulti([]string{"companies/1", "companies/2"})
		assert.NoError(t, err)
		assert.False(t, lazyOrders.IsValueCreated())

		orders := map[string]*Company{}
		err = lazyOrders.GetValue(orders)
		assert.NoError(t, err)
		assert.Equal(t, len(orders), 2)

		fmt.Printf("!!!orders:\n%#v\n", orders)

		company1 := orders["companies/1"]
		company2 := orders["companies/2"]

		assert.NotNil(t, company1)
		assert.NotNil(t, company2)

		assert.Equal(t, company1.ID, "companies/1")

		assert.Equal(t, company2.ID, "companies/2")

		lazyOrder, err = session.Advanced().Lazily().Load("companies/3")
		assert.NoError(t, err)
		assert.False(t, lazyOrder.IsValueCreated())

		err = lazyOrder.GetValue(&order)
		assert.NoError(t, err)
		assert.Equal(t, order.ID, "companies/3")

		load, err := session.Advanced().Lazily().LoadMulti([]string{"no_such_1", "no_such_2"})
		assert.NoError(t, err)
		missingItems := map[string]*Company{}
		err = load.GetValue(missingItems)
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
		fn1 := func() {
			assert.NotNil(t, company1Ref)
		}
		lazy1, err := query.LoadWithEval("companies/1", fn1, &company1Ref)
		assert.NoError(t, err)

		fn2 := func() {
			assert.NotNil(t, company2Ref)
		}
		lazy2, err := query.LoadWithEval("companies/2", fn2, &company2Ref)
		assert.NoError(t, err)

		assert.Nil(t, company1Ref)
		assert.Nil(t, company2Ref)

		_, err = session.Advanced().Eagerly().ExecuteAllPendingLazyOperations()
		assert.NoError(t, err)
		assert.Equal(t, company1Ref.ID, "companies/1")
		assert.Equal(t, company2Ref.ID, "companies/2")

		var c1, c2 *Company
		err = lazy1.GetValue(&c1)
		assert.NoError(t, err)
		err = lazy2.GetValue(&c2)
		assert.NoError(t, err)
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

		query := session.Advanced().Lazily()
		fn := func() {
			assert.NotNil(t, userRef)
		}
		lazy, err := query.LoadWithEval("users/1", fn, &userRef)
		assert.NoError(t, err)
		lazy.Value = &userRef

		assert.Nil(t, userRef)

		_, err = session.Advanced().Eagerly().ExecuteAllPendingLazyOperations()
		assert.NoError(t, err)
		assert.Equal(t, *userRef.LastName, "Oren")

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
		lazyUser, err := session.Advanced().Lazily().Load("users/1")
		assert.NoError(t, err)
		assert.False(t, lazyUser.IsValueCreated())

		var user *User
		err = lazyUser.GetValue(&user)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, user.ID, "users/1")

		session.Close()
	}

	{
		session := openSessionMust(t, store)
		lazyUser, err := session.Advanced().Lazily().Load("users/1")
		assert.NoError(t, err)
		assert.False(t, lazyUser.IsValueCreated())

		var user *User
		err = lazyUser.GetValue(&user)
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

		lazyLoad, err := session.Advanced().Lazily().LoadMulti([]string{"users/2", "users/3"})
		assert.NoError(t, err)

		//users2 := map[string]*User{}
		_, err = session.Advanced().Lazily().LoadMulti([]string{"users/1", "users/3"})
		assert.NoError(t, err)

		var u1, u2 *User
		err = session.Load(&u1, "users/2")
		assert.NoError(t, err)
		err = session.Load(&u2, "users/3")
		assert.NoError(t, err)

		_, err = session.Advanced().Eagerly().ExecuteAllPendingLazyOperations()
		assert.NoError(t, err)

		assert.True(t, session.Advanced().IsLoaded("users/1"))

		users := map[string]*User{}
		err = lazyLoad.GetValue(users)
		assert.NoError(t, err)
		assert.Equal(t, len(users), 2)

		oldRequestCount := session.Advanced().GetNumberOfRequests()

		lazyLoad, err = session.Advanced().Lazily().LoadMulti([]string{"users/3"})
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
