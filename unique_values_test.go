package ravendb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func uniqueValues_canReadNotExistingKey(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		op := NewGetCompareExchangeValueOperation(getTypeOf(0), "test")
		err = store.operations().send(op)
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
		op := NewGetCompareExchangeValueOperation(getTypeOf(0), "test")
		err = store.operations().send(op)
		assert.NoError(t, err)
		res := op.Command.Result
		assert.Nil(t, res)
	}
	{
		op := NewPutCompareExchangeValueOperation("test", 5, 0)
		err = store.operations().send(op)
		assert.NoError(t, err)
	}
	{
		op := NewGetCompareExchangeValueOperation(getTypeOf(0), "test")
		err = store.operations().send(op)
		assert.NoError(t, err)
		res := op.Command.Result
		assert.NotNil(t, res)
		v := res.getValue().(int)
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
		op := NewPutCompareExchangeValueOperation("test", "Karmel", 0)
		err = store.operations().send(op)
		assert.NoError(t, err)

		op2 := NewGetCompareExchangeValueOperation(getTypeOf(""), "test")
		err = store.operations().send(op2)
		assert.NoError(t, err)

		res := op2.Command.Result
		val := res.getValue().(string)
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

		op := NewPutCompareExchangeValueOperation("test", user1, 0)
		err = store.operations().send(op)
		assert.NoError(t, err)
		res := op.Command.Result

		user2 := NewUser()
		user2.setName("Karmel")

		op2 := NewPutCompareExchangeValueOperation("test2", user2, 0)
		err = store.operations().send(op2)
		assert.NoError(t, err)
		res2 := op2.Command.Result

		val := res.getValue().(*User)
		assert.Equal(t, *val.getName(), "Karmel")
		assert.True(t, res.isSuccessful())

		val2 := res2.getValue().(*User)
		assert.Equal(t, *val2.getName(), "Karmel")
		assert.True(t, res.isSuccessful())
	}
}

func uniqueValues_canListCompareExchange(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		user1 := NewUser()
		user1.setName("Karmel")
		op := NewPutCompareExchangeValueOperation("test", user1, 0)
		err = store.operations().send(op)
		assert.NoError(t, err)
		res1 := op.Command.Result
		val1 := res1.getValue().(*User)

		user2 := NewUser()
		user2.setName("Karmel")

		op2 := NewPutCompareExchangeValueOperation("test2", user2, 0)
		err = store.operations().send(op2)
		assert.NoError(t, err)
		res2 := op2.Command.Result
		val2 := res2.getValue().(*User)

		assert.Equal(t, *val1.getName(), "Karmel")
		assert.True(t, res1.isSuccessful())

		assert.Equal(t, *val2.getName(), "Karmel")
		assert.True(t, res2.isSuccessful())
	}
	{
		op := NewGetCompareExchangeValuesOperation(getTypeOf(&User{}), "test", -1, -1)
		err = store.operations().send(op)
		assert.NoError(t, err)
		values := op.Command.Result
		assert.Equal(t, len(values), 2)

		v := values["test"].getValue().(*User)
		assert.Equal(t, *v.getName(), "Karmel")

		v = values["test2"].getValue().(*User)
		assert.Equal(t, *v.getName(), "Karmel")

	}
}

func uniqueValues_canRemoveUnique(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		op := NewPutCompareExchangeValueOperation("test", "Karmel", 0)
		err = store.operations().send(op)
		assert.NoError(t, err)
		res := op.Command.Result
		val := res.getValue().(string)
		assert.Equal(t, val, "Karmel")
		assert.True(t, res.isSuccessful())
		{
			op := NewDeleteCompareExchangeValueOperation(getTypeOf(""), "test", res.getIndex())
			err = store.operations().send(op)
			assert.NoError(t, err)
			assert.True(t, res.isSuccessful())
		}
	}
}

func uniqueValues_removeUniqueFailed(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		op := NewPutCompareExchangeValueOperation("test", "Karmel", 0)
		err = store.operations().send(op)
		assert.NoError(t, err)
		res := op.Command.Result
		val := res.getValue().(string)
		assert.Equal(t, val, "Karmel")
		assert.True(t, res.isSuccessful())
	}
	{
		op := NewDeleteCompareExchangeValueOperation(getTypeOf(""), "test", 0)
		err = store.operations().send(op)
		assert.NoError(t, err)
		res := op.Command.Result
		assert.False(t, res.isSuccessful())
	}
	{
		op := NewGetCompareExchangeValueOperation(getTypeOf(""), "test")
		err = store.operations().send(op)
		assert.NoError(t, err)
		readValue := op.Command.Result
		val := readValue.getValue().(string)
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

		op := NewPutCompareExchangeValueOperation("test", user, 0)
		err = store.operations().send(op)
		assert.NoError(t, err)
		res := op.Command.Result

		op2 := NewPutCompareExchangeValueOperation("test", user2, 0)
		err = store.operations().send(op2)
		assert.NoError(t, err)
		res2 := op2.Command.Result

		assert.True(t, res.isSuccessful())
		assert.False(t, res2.isSuccessful())

		val := res.getValue().(*User)
		assert.Equal(t, *val.getName(), "Karmel")

		val2 := res2.getValue().(*User)
		assert.Equal(t, *val2.getName(), "Karmel")

		user3 := NewUser()
		user3.setName("Karmel2")

		op3 := NewPutCompareExchangeValueOperation("test", user3, res2.getIndex())
		err = store.operations().send(op3)
		assert.NoError(t, err)
		res2 = op3.Command.Result
		assert.True(t, res2.isSuccessful())
		val2 = res2.getValue().(*User)
		assert.Equal(t, *val2.getName(), "Karmel2")
	}
}

func uniqueValues_canGetIndexValue(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		user := NewUser()
		user.setName("Karmel")
		op := NewPutCompareExchangeValueOperation("test", user, 0)
		err = store.operations().send(op)
		assert.NoError(t, err)
	}
	{
		op := NewGetCompareExchangeValueOperation(getTypeOf(&User{}), "test")
		err = store.operations().send(op)
		assert.NoError(t, err)
		res := op.Command.Result
		val := res.getValue().(*User)
		assert.Equal(t, *val.getName(), "Karmel")

		user2 := NewUser()
		user2.setName("Karmel2")
		op2 := NewPutCompareExchangeValueOperation("test", user2, res.getIndex())
		err = store.operations().send(op2)
		assert.NoError(t, err)
		res2 := op2.Command.Result
		assert.True(t, res2.isSuccessful())
		val2 := res2.getValue().(*User)
		assert.Equal(t, *val2.getName(), "Karmel2")
	}
}

func TestUniqueValues(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer func() {
		r := recover()
		destroyDriver()
		if r != nil {
			panic(r)
		}
	}()

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
