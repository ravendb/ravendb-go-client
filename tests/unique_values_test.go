package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

func uniqueValues_canReadNotExistingKey(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		op := ravendb.NewGetCompareExchangeValueOperation(ravendb.GetTypeOf(0), "test")
		err = store.Operations().Send(op)
		assert.NoError(t, err)
		res := op.Command.Result
		assert.Nil(t, res)
	}
}

func uniqueValues_canWorkWithPrimitiveTypes(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		op := ravendb.NewGetCompareExchangeValueOperation(ravendb.GetTypeOf(0), "test")
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
		op := ravendb.NewGetCompareExchangeValueOperation(ravendb.GetTypeOf(0), "test")
		err = store.Operations().Send(op)
		assert.NoError(t, err)
		res := op.Command.Result
		assert.NotNil(t, res)
		v := res.GetValue().(int)
		assert.Equal(t, v, 5)
	}
}

func uniqueValues_canPutUniqueString(t *testing.T) {

	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		// Note: not sure why Java test opens a session
		_ = openSessionMust(t, store)
		op := ravendb.NewPutCompareExchangeValueOperation("test", "Karmel", 0)
		err = store.Operations().Send(op)
		assert.NoError(t, err)

		op2 := ravendb.NewGetCompareExchangeValueOperation(ravendb.GetTypeOf(""), "test")
		err = store.Operations().Send(op2)
		assert.NoError(t, err)

		res := op2.Command.Result
		val := res.GetValue().(string)
		assert.Equal(t, val, "Karmel")
	}
}

func uniqueValues_canPutMultiDifferentValues(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		user1 := NewUser()
		user1.setName("Karmel")

		op := ravendb.NewPutCompareExchangeValueOperation("test", user1, 0)
		err = store.Operations().Send(op)
		assert.NoError(t, err)
		res := op.Command.Result

		user2 := NewUser()
		user2.setName("Karmel")

		op2 := ravendb.NewPutCompareExchangeValueOperation("test2", user2, 0)
		err = store.Operations().Send(op2)
		assert.NoError(t, err)
		res2 := op2.Command.Result

		val := res.GetValue().(*User)
		assert.Equal(t, *val.GetName(), "Karmel")
		assert.True(t, res.IsSuccessful())

		val2 := res2.GetValue().(*User)
		assert.Equal(t, *val2.GetName(), "Karmel")
		assert.True(t, res.IsSuccessful())
	}
}

func uniqueValues_canListCompareExchange(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		user1 := NewUser()
		user1.setName("Karmel")
		op := ravendb.NewPutCompareExchangeValueOperation("test", user1, 0)
		err = store.Operations().Send(op)
		assert.NoError(t, err)
		res1 := op.Command.Result
		val1 := res1.GetValue().(*User)

		user2 := NewUser()
		user2.setName("Karmel")

		op2 := ravendb.NewPutCompareExchangeValueOperation("test2", user2, 0)
		err = store.Operations().Send(op2)
		assert.NoError(t, err)
		res2 := op2.Command.Result
		val2 := res2.GetValue().(*User)

		assert.Equal(t, *val1.GetName(), "Karmel")
		assert.True(t, res1.IsSuccessful())

		assert.Equal(t, *val2.GetName(), "Karmel")
		assert.True(t, res2.IsSuccessful())
	}
	{
		op := ravendb.NewGetCompareExchangeValuesOperation(ravendb.GetTypeOf(&User{}), "test", -1, -1)
		err = store.Operations().Send(op)
		assert.NoError(t, err)
		values := op.Command.Result
		assert.Equal(t, len(values), 2)

		v := values["test"].GetValue().(*User)
		assert.Equal(t, *v.GetName(), "Karmel")

		v = values["test2"].GetValue().(*User)
		assert.Equal(t, *v.GetName(), "Karmel")

	}
}

func uniqueValues_canRemoveUnique(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
			op := ravendb.NewDeleteCompareExchangeValueOperation(ravendb.GetTypeOf(""), "test", res.GetIndex())
			err = store.Operations().Send(op)
			assert.NoError(t, err)
			assert.True(t, res.IsSuccessful())
		}
	}
}

func uniqueValues_removeUniqueFailed(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
		op := ravendb.NewDeleteCompareExchangeValueOperation(ravendb.GetTypeOf(""), "test", 0)
		err = store.Operations().Send(op)
		assert.NoError(t, err)
		res := op.Command.Result
		assert.False(t, res.IsSuccessful())
	}
	{
		op := ravendb.NewGetCompareExchangeValueOperation(ravendb.GetTypeOf(""), "test")
		err = store.Operations().Send(op)
		assert.NoError(t, err)
		readValue := op.Command.Result
		val := readValue.GetValue().(string)
		assert.Equal(t, val, "Karmel")
	}
}

func uniqueValues_returnCurrentValueWhenPuttingConcurrently(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		user := NewUser()
		user.setName("Karmel")

		user2 := NewUser()
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
		assert.Equal(t, *val.GetName(), "Karmel")

		val2 := res2.GetValue().(*User)
		assert.Equal(t, *val2.GetName(), "Karmel")

		user3 := NewUser()
		user3.setName("Karmel2")

		op3 := ravendb.NewPutCompareExchangeValueOperation("test", user3, res2.GetIndex())
		err = store.Operations().Send(op3)
		assert.NoError(t, err)
		res2 = op3.Command.Result
		assert.True(t, res2.IsSuccessful())
		val2 = res2.GetValue().(*User)
		assert.Equal(t, *val2.GetName(), "Karmel2")
	}
}

func uniqueValues_canGetIndexValue(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		user := NewUser()
		user.setName("Karmel")
		op := ravendb.NewPutCompareExchangeValueOperation("test", user, 0)
		err = store.Operations().Send(op)
		assert.NoError(t, err)
	}
	{
		op := ravendb.NewGetCompareExchangeValueOperation(ravendb.GetTypeOf(&User{}), "test")
		err = store.Operations().Send(op)
		assert.NoError(t, err)
		res := op.Command.Result
		val := res.GetValue().(*User)
		assert.Equal(t, *val.GetName(), "Karmel")

		user2 := NewUser()
		user2.setName("Karmel2")
		op2 := ravendb.NewPutCompareExchangeValueOperation("test", user2, res.GetIndex())
		err = store.Operations().Send(op2)
		assert.NoError(t, err)
		res2 := op2.Command.Result
		assert.True(t, res2.IsSuccessful())
		val2 := res2.GetValue().(*User)
		assert.Equal(t, *val2.GetName(), "Karmel2")
	}
}

func TestUniqueValues(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	uniqueValues_removeUniqueFailed(t)
	uniqueValues_canGetIndexValue(t)
	uniqueValues_canRemoveUnique(t)
	uniqueValues_canWorkWithPrimitiveTypes(t)
	uniqueValues_canReadNotExistingKey(t)
	uniqueValues_canPutMultiDifferentValues(t)
	uniqueValues_canPutUniqueString(t)
	uniqueValues_canListCompareExchange(t)
	uniqueValues_returnCurrentValueWhenPuttingConcurrently(t)
}
