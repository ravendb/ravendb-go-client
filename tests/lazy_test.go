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
	}

	{
		session := openSessionMust(t, store)
		query := session.Advanced().Lazily()
		var order *Company
		lazyOrder := query.Load(&order, "companies/1", nil)

		assert.False(t, lazyOrder.IsValueCreated())
		err = lazyOrder.GetValue()
		assert.NoError(t, err)
		assert.Equal(t, order.ID, "companies/1")

		orders := map[string]*Company{}
		lazyOrders := session.Advanced().Lazily().LoadMulti(orders, []string{"companies/1", "companies/2"}, nil)
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

		lazyOrder = session.Advanced().Lazily().Load(&order, "companies/3", nil)
		assert.False(t, lazyOrder.IsValueCreated())

		err = lazyOrder.GetValue()
		assert.NoError(t, err)
		assert.Equal(t, order.ID, "companies/3")

		missingItems := map[string]*Company{}
		load := session.Advanced().Lazily().LoadMulti(missingItems, []string{"no_such_1", "no_such_2"}, nil)
		err = load.GetValue()
		assert.NoError(t, err)
		assert.Equal(t, 2, len(missingItems))

		assert.Nil(t, missingItems["no_such_1"])
		assert.Nil(t, missingItems["no_such_2"])
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
	}

	{
		session := openSessionMust(t, store)
		var company1Ref *Company
		var company2Ref *Company
		query := session.Advanced().Lazily()
		var c1, c2 *Company
		query.Load(&c1, "companies/1", func(v interface{}) {
			c := v.(*Company)
			company1Ref = c
		})

		query.Load(&c2, "companies/2", func(v interface{}) {
			c := v.(*Company)
			company2Ref = c
		})

		assert.Nil(t, company1Ref)
		assert.Nil(t, company2Ref)

		_, err = session.Advanced().Eagerly().ExecuteAllPendingLazyOperations()
		assert.NoError(t, err)
		assert.Equal(t, company1Ref.ID, "companies/1")
		assert.Equal(t, company2Ref.ID, "companies/2")

		assert.Equal(t, c1, company1Ref)
		assert.Equal(t, c2, company2Ref)
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
	}

	{
		session := openSessionMust(t, store)

		var userRef *User
		var user *User

		query := session.Advanced().Lazily()
		query.Load(&user, "users/1", func(v interface{}) {
			userRef = v.(*User)
		})

		assert.Nil(t, userRef)

		_, err = session.Advanced().Eagerly().ExecuteAllPendingLazyOperations()
		assert.NoError(t, err)
		assert.Equal(t, *userRef.LastName, "Oren")
		assert.Equal(t, user, userRef)
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
	}

	{
		session := openSessionMust(t, store)
		var user *User
		lazyUser := session.Advanced().Lazily().Load(&user, "users/1", nil)
		assert.False(t, lazyUser.IsValueCreated())

		err = lazyUser.GetValue()
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, user.ID, "users/1")
	}

	{
		session := openSessionMust(t, store)
		var user *User
		lazyUser := session.Advanced().Lazily().Load(&user, "users/1", nil)
		assert.False(t, lazyUser.IsValueCreated())

		err = lazyUser.GetValue()
		assert.NoError(t, err)
		assert.Equal(t, *user.LastName, "Oren")
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
}
