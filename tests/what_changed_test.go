package tests

import (
	"fmt"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func whatChanged_whatChangedNewField(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		basicName := &BasicName{}
		basicName.Name = "Toli"
		err = newSession.StoreWithID(basicName, "users/1")
		assert.NoError(t, err)

		changes, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 1)
		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}

	{
		var user *NameAndAge
		newSession := openSessionMust(t, store)
		err = newSession.Load(&user, "users/1")
		assert.NoError(t, err)
		user.Age = 5

		changesMap, _ := newSession.Advanced().WhatChanged()
		changes := changesMap["users/1"]
		assert.Equal(t, len(changes), 1)

		{
			change := changes[0]
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_NEW_FIELD)
			err = newSession.SaveChanges()
			assert.NoError(t, err)
		}
		newSession.Close()
	}
}

func whatChanged_whatChangedRemovedField(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		nameAndAge := &NameAndAge{}
		nameAndAge.Age = 5
		nameAndAge.Name = "Toli"

		err = newSession.StoreWithID(nameAndAge, "users/1")
		assert.NoError(t, err)

		changes, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 1)

		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}

	{
		newSession := openSessionMust(t, store)
		var unused *BasicAge
		err = newSession.Load(&unused, "users/1")
		assert.NoError(t, err)

		changesMap, _ := newSession.Advanced().WhatChanged()
		changes := changesMap["users/1"]
		assert.Equal(t, len(changes), 1)

		{
			change := changes[0]
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_REMOVED_FIELD)
		}

		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}
}

func whatChanged_whatChangedChangeField(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		basicAge := &BasicAge{}
		basicAge.Age = 5
		err = newSession.StoreWithID(basicAge, "users/1")
		assert.NoError(t, err)

		changes, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 1)

		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}

	{
		newSession := openSessionMust(t, store)
		var unused *Int
		err = newSession.Load(&unused, "users/1")
		assert.NoError(t, err)
		changesMap, _ := newSession.Advanced().WhatChanged()
		changes := changesMap["users/1"]
		assert.Equal(t, len(changes), 2)

		{
			change := changes[0]
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_REMOVED_FIELD)
		}

		{
			change := changes[1]
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_NEW_FIELD)
		}

		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}
}

func whatChanged_whatChangedArrayValueChanged(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		arr := &Arr{}
		arr.Array = []interface{}{"a", 1, "b"}

		err = session.StoreWithID(arr, "users/1")
		assert.NoError(t, err)
		changesMap, _ := session.Advanced().WhatChanged()
		assert.Equal(t, len(changesMap), 1)

		changes := changesMap["users/1"]
		assert.Equal(t, len(changes), 1)

		{
			change := changes[0]
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_DOCUMENT_ADDED)
			err = session.SaveChanges()
			assert.NoError(t, err)
		}

		{
			newSession := openSessionMust(t, store)
			var arr *Arr
			err = newSession.Load(&arr, "users/1")
			assert.NoError(t, err)

			arr.Array = []interface{}{"a", 2, "c"}

			changesMap, _ := newSession.Advanced().WhatChanged()
			assert.Equal(t, len(changesMap), 1)

			changes := changesMap["users/1"]
			assert.Equal(t, len(changes), 2)

			{
				change := changes[0]
				assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_ARRAY_VALUE_CHANGED)
				oldValue := change.GetFieldOldValue()
				assert.Equal(t, oldValue, 1.0)
				newValue := change.GetFieldNewValue()
				assert.Equal(t, newValue, 2.0)
			}

			{
				change := changes[1]
				assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_ARRAY_VALUE_CHANGED)
				oldValueStr := fmt.Sprintf("%#v", change.GetFieldOldValue())
				assert.Equal(t, oldValueStr, "\"b\"")
				newValueStr := fmt.Sprintf("%#v", change.GetFieldNewValue())
				assert.Equal(t, newValueStr, "\"c\"")
			}
			newSession.Close()
		}
		session.Close()
	}
}

func whatChanged_what_Changed_Array_Value_Added(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		arr := &Arr{}
		arr.Array = []interface{}{"a", 1, "b"}
		err = newSession.StoreWithID(arr, "arr/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}

	{
		newSession := openSessionMust(t, store)
		var arr *Arr
		err = newSession.Load(&arr, "arr/1")
		assert.NoError(t, err)

		arr.Array = []interface{}{"a", 1, "b", "c", 2}

		changesMap, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(changesMap), 1)
		changes := changesMap["arr/1"]
		assert.Equal(t, len(changes), 2)

		{
			change := changes[0]
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_ARRAY_VALUE_ADDED)
			newValStr := fmt.Sprintf("%#v", change.GetFieldNewValue())
			assert.Equal(t, newValStr, "\"c\"")
			assert.Nil(t, change.GetFieldOldValue())
		}
		{
			change := changes[1]
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_ARRAY_VALUE_ADDED)
			assert.Equal(t, change.GetFieldNewValue(), float64(2))
			assert.Nil(t, change.GetFieldOldValue())
		}
		newSession.Close()
	}
}

func whatChanged_what_Changed_Array_Value_Removed(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		arr := &Arr{}
		arr.Array = []interface{}{"a", 1, "b"}
		err = newSession.StoreWithID(arr, "arr/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}

	{
		newSession := openSessionMust(t, store)
		var arr *Arr
		err = newSession.Load(&arr, "arr/1")
		assert.NoError(t, err)

		arr.Array = []interface{}{"a"}

		changesMap, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(changesMap), 1)
		changes := changesMap["arr/1"]
		assert.Equal(t, len(changes), 2)

		{
			change := changes[0]
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_ARRAY_VALUE_REMOVED)
			assert.Equal(t, change.GetFieldOldValue(), float64(1))
			assert.Nil(t, change.GetFieldNewValue())
		}

		{
			change := changes[1]
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_ARRAY_VALUE_REMOVED)

			oldValStr := fmt.Sprintf("%#v", change.GetFieldOldValue())
			assert.Equal(t, oldValStr, "\"b\"")
			assert.Nil(t, change.GetFieldNewValue())
		}
		newSession.Close()
	}
}

func whatChanged_ravenDB_8169(t *testing.T, driver *RavenTestDriver) {
	//Test that when old and new values are of different type
	//but have the same value, we consider them unchanged

	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)

		anInt := &Int{}
		anInt.Number = 1

		err = newSession.StoreWithID(anInt, "num/1")
		assert.NoError(t, err)

		aDouble := &Double{}
		aDouble.Number = 2.0
		err = newSession.StoreWithID(aDouble, "num/2")
		assert.NoError(t, err)

		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}

	{
		newSession := openSessionMust(t, store)
		var unused *Double
		err = newSession.Load(&unused, "num/1")
		assert.NoError(t, err)

		changes, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 0)
		newSession.Close()
	}

	{
		newSession := openSessionMust(t, store)
		var unused *Int
		err = newSession.Load(&unused, "num/2")
		assert.NoError(t, err)

		changes, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 0)
		newSession.Close()
	}
}

func whatChanged_whatChanged_should_be_idempotent_operation(t *testing.T, driver *RavenTestDriver) {
	//RavenDB-9150
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user1 := &User{}
		user1.setName("user1")

		user2 := &User{}
		user2.setName("user2")
		user2.Age = 1

		user3 := &User{}
		user3.setName("user3")
		user3.Age = 1

		err = session.StoreWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.StoreWithID(user2, "users/2")
		assert.NoError(t, err)
		err = session.StoreWithID(user3, "users/3")
		assert.NoError(t, err)

		changes, _ := session.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 3)

		err = session.SaveChanges()
		assert.NoError(t, err)

		user1 = nil
		err = session.Load(&user1, "users/1")
		assert.NoError(t, err)

		user2 = nil
		err = session.Load(&user2, "users/2")
		assert.NoError(t, err)

		user1.Age = 10
		err = session.DeleteEntity(&user2)
		assert.NoError(t, err)

		changes, _ = session.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 2)
		changes, _ = session.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 2)
		session.Close()
	}
}

type BasicName struct {
	Name string
}

type NameAndAge struct {
	Name string
	Age  int
}

type BasicAge struct {
	Age int
}

type Int struct {
	Number int
}

type Double struct {
	Number float64
}

type Arr struct {
	Array []interface{}
}

func TestWhatChanged(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	whatChanged_what_Changed_Array_Value_Removed(t, driver)
	whatChanged_whatChangedNewField(t, driver)
	whatChanged_what_Changed_Array_Value_Added(t, driver)
	whatChanged_whatChangedChangeField(t, driver)
	whatChanged_whatChangedArrayValueChanged(t, driver)
	whatChanged_ravenDB_8169(t, driver)
	whatChanged_whatChangedRemovedField(t, driver)
	whatChanged_whatChanged_should_be_idempotent_operation(t, driver)
}
