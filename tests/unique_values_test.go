package tests

import (
	"reflect"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func uniqueValues_canReadNotExistingKey(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		op := ravendb.NewGetCompareExchangeValueOperation(reflect.TypeOf(0), "test")
		err = store.Operations().Send(op)
		assert.NoError(t, err)
		res := op.Command.Result
		assert.Nil(t, res)
	}
}

func uniqueValues_canWorkWithPrimitiveTypes(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		op := ravendb.NewGetCompareExchangeValueOperation(reflect.TypeOf(0), "test")
		err = store.Operations().Send(op)
		assert.NoError(t, err)
		res := op.Command.Result
		assert.Nil(t, res)
	}
	{
		op := ravendb.NewPutCompareExchangeValueOperation("test", 5, 0)
		err = store.Operations().Send(op)
		assert.NoError(t, err)
	}
	{
		op := ravendb.NewGetCompareExchangeValueOperation(reflect.TypeOf(0), "test")
		err = store.Operations().Send(op)
		assert.NoError(t, err)
		res := op.Command.Result
		assert.NotNil(t, res)
		v := res.GetValue().(int)
		assert.Equal(t, v, 5)
	}
}

func uniqueValues_canPutUniqueString(t *testing.T, driver *RavenTestDriver) {

	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		// Note: not sure why Java test opens a session
		_ = openSessionMust(t, store)
		op := ravendb.NewPutCompareExchangeValueOperation("test", "Karmel", 0)
		err = store.Operations().Send(op)
		assert.NoError(t, err)

		op2 := ravendb.NewGetCompareExchangeValueOperation(reflect.TypeOf(""), "test")
		err = store.Operations().Send(op2)
		assert.NoError(t, err)

		res := op2.Command.Result
		val := res.GetValue().(string)
		assert.Equal(t, val, "Karmel")
	}
}

func uniqueValues_canPutMultiDifferentValues(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		user1 := &User{}
		user1.setName("Karmel")

		op := ravendb.NewPutCompareExchangeValueOperation("test", user1, 0)
		err = store.Operations().Send(op)
		assert.NoError(t, err)
		res := op.Command.Result

		user2 := &User{}
		user2.setName("Karmel")

		op2 := ravendb.NewPutCompareExchangeValueOperation("test2", user2, 0)
		err = store.Operations().Send(op2)
		assert.NoError(t, err)
		res2 := op2.Command.Result

		val := res.GetValue().(*User)
		assert.Equal(t, *val.Name, "Karmel")
		assert.True(t, res.IsSuccessful())

		val2 := res2.GetValue().(*User)
		assert.Equal(t, *val2.Name, "Karmel")
		assert.True(t, res.IsSuccessful())
	}
}

func uniqueValues_canListCompareExchange(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		user1 := &User{}
		user1.setName("Karmel")
		op := ravendb.NewPutCompareExchangeValueOperation("test", user1, 0)
		err = store.Operations().Send(op)
		assert.NoError(t, err)
		res1 := op.Command.Result
		val1 := res1.GetValue().(*User)

		user2 := &User{}
		user2.setName("Karmel")

		op2 := ravendb.NewPutCompareExchangeValueOperation("test2", user2, 0)
		err = store.Operations().Send(op2)
		assert.NoError(t, err)
		res2 := op2.Command.Result
		val2 := res2.GetValue().(*User)

		assert.Equal(t, *val1.Name, "Karmel")
		assert.True(t, res1.IsSuccessful())

		assert.Equal(t, *val2.Name, "Karmel")
		assert.True(t, res2.IsSuccessful())
	}
	{
		op := ravendb.NewGetCompareExchangeValuesOperation(reflect.TypeOf(&User{}), "test", -1, -1)
		err = store.Operations().Send(op)
		assert.NoError(t, err)
		values := op.Command.Result
		assert.Equal(t, len(values), 2)

		v := values["test"].GetValue().(*User)
		assert.Equal(t, *v.Name, "Karmel")

		v = values["test2"].GetValue().(*User)
		assert.Equal(t, *v.Name, "Karmel")

	}
}

func uniqueValues_canRemoveUnique(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		op := ravendb.NewPutCompareExchangeValueOperation("test", "Karmel", 0)
		err = store.Operations().Send(op)
		assert.NoError(t, err)
		res := op.Command.Result
		val := res.GetValue().(string)
		assert.Equal(t, val, "Karmel")
		assert.True(t, res.IsSuccessful())
		{
			op := ravendb.NewDeleteCompareExchangeValueOperation(reflect.TypeOf(""), "test", res.GetIndex())
			err = store.Operations().Send(op)
			assert.NoError(t, err)
			assert.True(t, res.IsSuccessful())
		}
	}
}

func uniqueValues_removeUniqueFailed(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		op := ravendb.NewPutCompareExchangeValueOperation("test", "Karmel", 0)
		err = store.Operations().Send(op)
		assert.NoError(t, err)
		res := op.Command.Result
		val := res.GetValue().(string)
		assert.Equal(t, val, "Karmel")
		assert.True(t, res.IsSuccessful())
	}
	{
		op := ravendb.NewDeleteCompareExchangeValueOperation(reflect.TypeOf(""), "test", 0)
		err = store.Operations().Send(op)
		assert.NoError(t, err)
		res := op.Command.Result
		assert.False(t, res.IsSuccessful())
	}
	{
		op := ravendb.NewGetCompareExchangeValueOperation(reflect.TypeOf(""), "test")
		err = store.Operations().Send(op)
		assert.NoError(t, err)
		readValue := op.Command.Result
		val := readValue.GetValue().(string)
		assert.Equal(t, val, "Karmel")
	}
}

func uniqueValues_returnCurrentValueWhenPuttingConcurrently(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		user := &User{}
		user.setName("Karmel")

		user2 := &User{}
		user2.setName("Karmel2")

		op := ravendb.NewPutCompareExchangeValueOperation("test", user, 0)
		err = store.Operations().Send(op)
		assert.NoError(t, err)
		res := op.Command.Result

		op2 := ravendb.NewPutCompareExchangeValueOperation("test", user2, 0)
		err = store.Operations().Send(op2)
		assert.NoError(t, err)
		res2 := op2.Command.Result

		assert.True(t, res.IsSuccessful())
		assert.False(t, res2.IsSuccessful())

		val := res.GetValue().(*User)
		assert.Equal(t, *val.Name, "Karmel")

		val2 := res2.GetValue().(*User)
		assert.Equal(t, *val2.Name, "Karmel")

		user3 := &User{}
		user3.setName("Karmel2")

		op3 := ravendb.NewPutCompareExchangeValueOperation("test", user3, res2.GetIndex())
		err = store.Operations().Send(op3)
		assert.NoError(t, err)
		res2 = op3.Command.Result
		assert.True(t, res2.IsSuccessful())
		val2 = res2.GetValue().(*User)
		assert.Equal(t, *val2.Name, "Karmel2")
	}
}

func uniqueValues_canGetIndexValue(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		user := &User{}
		user.setName("Karmel")
		op := ravendb.NewPutCompareExchangeValueOperation("test", user, 0)
		err = store.Operations().Send(op)
		assert.NoError(t, err)
	}
	{
		op := ravendb.NewGetCompareExchangeValueOperation(reflect.TypeOf(&User{}), "test")
		err = store.Operations().Send(op)
		assert.NoError(t, err)
		res := op.Command.Result
		val := res.GetValue().(*User)
		assert.Equal(t, *val.Name, "Karmel")

		user2 := &User{}
		user2.setName("Karmel2")
		op2 := ravendb.NewPutCompareExchangeValueOperation("test", user2, res.GetIndex())
		err = store.Operations().Send(op2)
		assert.NoError(t, err)
		res2 := op2.Command.Result
		assert.True(t, res2.IsSuccessful())
		val2 := res2.GetValue().(*User)
		assert.Equal(t, *val2.Name, "Karmel2")
	}
}

func TestUniqueValues(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	uniqueValues_removeUniqueFailed(t, driver)
	uniqueValues_canGetIndexValue(t, driver)
	uniqueValues_canRemoveUnique(t, driver)
	uniqueValues_canWorkWithPrimitiveTypes(t, driver)
	uniqueValues_canReadNotExistingKey(t, driver)
	uniqueValues_canPutMultiDifferentValues(t, driver)
	uniqueValues_canPutUniqueString(t, driver)
	uniqueValues_canListCompareExchange(t, driver)
	uniqueValues_returnCurrentValueWhenPuttingConcurrently(t, driver)
}
