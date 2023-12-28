package tests

import (
	"reflect"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func uniqueValuesCanReadNotExistingKey(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		op, err := ravendb.NewGetCompareExchangeValueOperation(reflect.TypeOf(0), "test")
		assert.NoError(t, err)
		err = store.Operations().Send(op, nil)
		assert.NoError(t, err)
		res := op.Command.Result
		assert.Nil(t, res)
	}
}

func uniqueValuesCanWorkWithPrimitiveTypes(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		op, err := ravendb.NewGetCompareExchangeValueOperation(reflect.TypeOf(0), "test")
		assert.NoError(t, err)
		err = store.Operations().Send(op, nil)
		assert.NoError(t, err)
		res := op.Command.Result
		assert.Nil(t, res)
	}
	{
		op, err := ravendb.NewPutCompareExchangeValueOperation("test", 5, 0)
		assert.NoError(t, err)
		err = store.Operations().Send(op, nil)
		assert.NoError(t, err)
	}
	{
		op, err := ravendb.NewGetCompareExchangeValueOperation(reflect.TypeOf(0), "test")
		assert.NoError(t, err)
		err = store.Operations().Send(op, nil)
		assert.NoError(t, err)
		res := op.Command.Result
		assert.NotNil(t, res)
		v := res.Value.(int)
		assert.Equal(t, v, 5)
	}
}

func uniqueValuesCanPutUniqueString(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		// Note: not sure why Java test opens a session
		_ = openSessionMust(t, store)
		op, err := ravendb.NewPutCompareExchangeValueOperation("test", "Karmel", 0)
		assert.NoError(t, err)
		err = store.Operations().Send(op, nil)
		assert.NoError(t, err)

		op2, err := ravendb.NewGetCompareExchangeValueOperation(reflect.TypeOf(""), "test")
		assert.NoError(t, err)
		err = store.Operations().Send(op2, nil)
		assert.NoError(t, err)

		res := op2.Command.Result
		val := res.Value.(string)
		assert.Equal(t, val, "Karmel")
	}
}

func uniqueValuesCanPutMultiDifferentValues(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		user1 := &User{}
		user1.setName("Karmel")

		op, err := ravendb.NewPutCompareExchangeValueOperation("test", user1, 0)
		assert.NoError(t, err)
		err = store.Operations().Send(op, nil)
		assert.NoError(t, err)
		res := op.Command.Result

		user2 := &User{}
		user2.setName("Karmel")

		op2, err := ravendb.NewPutCompareExchangeValueOperation("test2", user2, 0)
		assert.NoError(t, err)
		err = store.Operations().Send(op2, nil)
		assert.NoError(t, err)
		res2 := op2.Command.Result

		val := res.Value.(*User)
		assert.Equal(t, *val.Name, "Karmel")
		assert.True(t, res.IsSuccessful)

		val2 := res2.Value.(*User)
		assert.Equal(t, *val2.Name, "Karmel")
		assert.True(t, res.IsSuccessful)
	}
}

func uniqueValuesCanListCompareExchange(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		user1 := &User{}
		user1.setName("Karmel")
		op, err := ravendb.NewPutCompareExchangeValueOperation("test", user1, 0)
		assert.NoError(t, err)
		err = store.Operations().Send(op, nil)
		assert.NoError(t, err)
		res1 := op.Command.Result
		val1 := res1.Value.(*User)

		user2 := &User{}
		user2.setName("Karmel")

		op2, err := ravendb.NewPutCompareExchangeValueOperation("test2", user2, 0)
		assert.NoError(t, err)
		err = store.Operations().Send(op2, nil)
		assert.NoError(t, err)
		res2 := op2.Command.Result
		val2 := res2.Value.(*User)

		assert.Equal(t, *val1.Name, "Karmel")
		assert.True(t, res1.IsSuccessful)

		assert.Equal(t, *val2.Name, "Karmel")
		assert.True(t, res2.IsSuccessful)
	}
	{
		op, err := ravendb.NewGetCompareExchangeValuesOperation(reflect.TypeOf(&User{}), "test", -1, -1)
		assert.NoError(t, err)
		err = store.Operations().Send(op, nil)
		assert.NoError(t, err)
		values := op.Command.Result
		assert.Equal(t, len(values), 2)

		v := values["test"].Value.(*User)
		assert.Equal(t, *v.Name, "Karmel")

		v = values["test2"].Value.(*User)
		assert.Equal(t, *v.Name, "Karmel")
	}

	{
		op, err := ravendb.NewGetCompareExchangeValueOperation(reflect.TypeOf(&User{}), "test")
		assert.NoError(t, err)
		err = store.Operations().Send(op, nil)
		assert.NoError(t, err)
		v, ok := op.Command.Result.Value.(*User)
		assert.True(t, ok)
		assert.Equal(t, *v.Name, "Karmel")

	}
}

func uniqueValuesCanRemoveUnique(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		op, err := ravendb.NewPutCompareExchangeValueOperation("test", "Karmel", 0)
		assert.NoError(t, err)
		err = store.Operations().Send(op, nil)
		assert.NoError(t, err)
		res := op.Command.Result
		val := res.Value.(string)
		assert.Equal(t, val, "Karmel")
		assert.True(t, res.IsSuccessful)
		{
			op, err := ravendb.NewDeleteCompareExchangeValueOperation(reflect.TypeOf(""), "test", res.Index)
			assert.NoError(t, err)
			err = store.Operations().Send(op, nil)
			assert.NoError(t, err)
			assert.True(t, res.IsSuccessful)
		}
	}
}

func uniqueValuesRemoveUniqueFailed(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		op, err := ravendb.NewPutCompareExchangeValueOperation("test", "Karmel", 0)
		assert.NoError(t, err)
		err = store.Operations().Send(op, nil)
		assert.NoError(t, err)
		res := op.Command.Result
		val := res.Value.(string)
		assert.Equal(t, val, "Karmel")
		assert.True(t, res.IsSuccessful)
	}
	{
		op, err := ravendb.NewDeleteCompareExchangeValueOperation(reflect.TypeOf(""), "test", 0)
		assert.NoError(t, err)
		err = store.Operations().Send(op, nil)
		assert.NoError(t, err)
		res := op.Command.Result
		assert.False(t, res.IsSuccessful)
	}
	{
		op, err := ravendb.NewGetCompareExchangeValueOperation(reflect.TypeOf(""), "test")
		assert.NoError(t, err)
		err = store.Operations().Send(op, nil)
		assert.NoError(t, err)
		readValue := op.Command.Result
		val := readValue.Value.(string)
		assert.Equal(t, val, "Karmel")
	}
}

func uniqueValuesReturnCurrentValueWhenPuttingConcurrently(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		user := &User{}
		user.setName("Karmel")

		user2 := &User{}
		user2.setName("Karmel2")

		op, err := ravendb.NewPutCompareExchangeValueOperation("test", user, 0)
		assert.NoError(t, err)
		err = store.Operations().Send(op, nil)
		assert.NoError(t, err)
		res := op.Command.Result

		op2, err := ravendb.NewPutCompareExchangeValueOperation("test", user2, 0)
		assert.NoError(t, err)
		err = store.Operations().Send(op2, nil)
		assert.NoError(t, err)
		res2 := op2.Command.Result

		assert.True(t, res.IsSuccessful)
		assert.False(t, res2.IsSuccessful)

		val := res.Value.(*User)
		assert.Equal(t, *val.Name, "Karmel")

		val2 := res2.Value.(*User)
		assert.Equal(t, *val2.Name, "Karmel")

		user3 := &User{}
		user3.setName("Karmel2")

		op3, err := ravendb.NewPutCompareExchangeValueOperation("test", user3, res2.Index)
		assert.NoError(t, err)
		err = store.Operations().Send(op3, nil)
		assert.NoError(t, err)
		res2 = op3.Command.Result
		assert.True(t, res2.IsSuccessful)
		val2 = res2.Value.(*User)
		assert.Equal(t, *val2.Name, "Karmel2")
	}
}

func uniqueValuesCanGetIndexValue(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		user := &User{}
		user.setName("Karmel")
		op, err := ravendb.NewPutCompareExchangeValueOperation("test", user, 0)
		assert.NoError(t, err)
		err = store.Operations().Send(op, nil)
		assert.NoError(t, err)
	}
	{
		op, err := ravendb.NewGetCompareExchangeValueOperation(reflect.TypeOf(&User{}), "test")
		assert.NoError(t, err)
		err = store.Operations().Send(op, nil)
		assert.NoError(t, err)
		res := op.Command.Result
		val := res.Value.(*User)
		assert.Equal(t, *val.Name, "Karmel")

		user2 := &User{}
		user2.setName("Karmel2")
		op2, err := ravendb.NewPutCompareExchangeValueOperation("test", user2, res.Index)
		assert.NoError(t, err)
		err = store.Operations().Send(op2, nil)
		assert.NoError(t, err)
		res2 := op2.Command.Result
		assert.True(t, res2.IsSuccessful)
		val2 := res2.Value.(*User)
		assert.Equal(t, *val2.Name, "Karmel2")
	}
}

func TestUniqueValues(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	uniqueValuesRemoveUniqueFailed(t, driver)
	uniqueValuesCanGetIndexValue(t, driver)
	uniqueValuesCanRemoveUnique(t, driver)
	uniqueValuesCanWorkWithPrimitiveTypes(t, driver)
	uniqueValuesCanReadNotExistingKey(t, driver)
	uniqueValuesCanPutMultiDifferentValues(t, driver)
	uniqueValuesCanPutUniqueString(t, driver)
	uniqueValuesCanListCompareExchange(t, driver)
	uniqueValuesReturnCurrentValueWhenPuttingConcurrently(t, driver)
}
