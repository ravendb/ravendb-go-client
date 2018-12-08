package tests

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func lazy_canLazilyLoadEntity(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
		lazyOrder := query.Load(reflect.TypeOf(&Company{}), "companies/1", nil)

		assert.False(t, lazyOrder.IsValueCreated())
		orderI, err := lazyOrder.GetValue()
		assert.NoError(t, err)
		order := orderI.(*Company)
		assert.Equal(t, order.ID, "companies/1")

		lazyOrders := session.Advanced().Lazily().LoadMulti(reflect.TypeOf(&Company{}), []string{"companies/1", "companies/2"}, nil)
		assert.False(t, lazyOrders.IsValueCreated())

		ordersI, err := lazyOrders.GetValue()
		assert.NoError(t, err)
		orders := ordersI.(map[string](*Company))
		assert.Equal(t, len(orders), 2)

		company1 := orders["companies/1"]
		company2 := orders["companies/2"]

		assert.NotNil(t, company1)
		assert.NotNil(t, company2)

		assert.Equal(t, company1.ID, "companies/1")

		assert.Equal(t, company2.ID, "companies/2")

		lazyOrder = session.Advanced().Lazily().Load(reflect.TypeOf(&Company{}), "companies/3", nil)
		assert.False(t, lazyOrder.IsValueCreated())

		orderI, err = lazyOrder.GetValue()
		assert.NoError(t, err)
		order = orderI.(*Company)
		assert.Equal(t, order.ID, "companies/3")

		load := session.Advanced().Lazily().LoadMulti(reflect.TypeOf(&Company{}), []string{"no_such_1", "no_such_2"}, nil)
		missingItemsI, err := load.GetValue()
		assert.NoError(t, err)
		missingItems := missingItemsI.(map[string]*Company)

		assert.Nil(t, missingItems["no_such_1"])
		assert.Nil(t, missingItems["no_such_2"])
	}
}

func lazy_canExecuteAllPendingLazyOperations(t *testing.T) {}
func lazy_withQueuedActions_Load(t *testing.T)             {}
func lazy_canUseCacheWhenLazyLoading(t *testing.T)         {}

func TestLazy(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	lazy_canExecuteAllPendingLazyOperations(t)
	lazy_canLazilyLoadEntity(t)
	lazy_canUseCacheWhenLazyLoading(t)
	lazy_withQueuedActions_Load(t)
}
