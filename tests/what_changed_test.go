package tests

import (
	"fmt"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func whatChanged_whatChangedNewField(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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

		changes, _ := newSession.Advanced().WhatChanged()
		change := changes["users/1"]
		assert.Equal(t, len(change), 1)

		{
			change := change[0]
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_NEW_FIELD)
			err = newSession.SaveChanges()
			assert.NoError(t, err)
		}
		newSession.Close()
	}
}

func whatChanged_whatChangedRemovedField(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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

		changes, _ := newSession.Advanced().WhatChanged()
		change := changes["users/1"]
		assert.Equal(t, len(change), 1)

		{
			change := change[0]
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_REMOVED_FIELD)
		}

		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}
}

func whatChanged_whatChangedChangeField(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
		changes, _ := newSession.Advanced().WhatChanged()
		change := changes["users/1"]
		assert.Equal(t, len(change), 2)

		{
			change := change[0]
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_REMOVED_FIELD)
		}

		{
			change := change[1]
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_NEW_FIELD)
		}

		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}
}

func whatChanged_whatChangedArrayValueChanged(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		arr := &Arr{}
		arr.Array = []ravendb.Object{"a", 1, "b"}

		err = newSession.StoreWithID(arr, "users/1")
		assert.NoError(t, err)
		changes, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 1)

		change := changes["users/1"]
		assert.Equal(t, len(change), 1)

		{
			change := change[0]
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_DOCUMENT_ADDED)
			err = newSession.SaveChanges()
			assert.NoError(t, err)
		}

		{
			newSession := openSessionMust(t, store)
			var arr *Arr
			err = newSession.Load(&arr, "users/1")
			assert.NoError(t, err)

			arr.Array = []ravendb.Object{"a", 2, "c"}

			changes, _ := newSession.Advanced().WhatChanged()
			assert.Equal(t, len(changes), 1)

			change := changes["users/1"]
			assert.Equal(t, len(change), 2)

			{
				change := change[0]
				assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_ARRAY_VALUE_CHANGED)
				oldValue := change.GetFieldOldValue()
				assert.Equal(t, oldValue, 1.0)
				newValue := change.GetFieldNewValue()
				assert.Equal(t, newValue, 2.0)
			}

			{
				change := change[1]
				assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_ARRAY_VALUE_CHANGED)
				oldValueStr := fmt.Sprintf("%#v", change.GetFieldOldValue())
				assert.Equal(t, oldValueStr, "\"b\"")
				newValueStr := fmt.Sprintf("%#v", change.GetFieldNewValue())
				assert.Equal(t, newValueStr, "\"c\"")
			}
			newSession.Close()
		}
		newSession.Close()
	}
}

func whatChanged_what_Changed_Array_Value_Added(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		arr := &Arr{}
		arr.Array = []ravendb.Object{"a", 1, "b"}
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

		arr.Array = []ravendb.Object{"a", 1, "b", "c", 2}

		changes, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 1)
		change := changes["arr/1"]
		assert.Equal(t, len(change), 2)

		{
			change := change[0]
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_ARRAY_VALUE_ADDED)
			newValStr := fmt.Sprintf("%#v", change.GetFieldNewValue())
			assert.Equal(t, newValStr, "\"c\"")
			assert.Nil(t, change.GetFieldOldValue())
		}
		{
			change := change[1]
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_ARRAY_VALUE_ADDED)
			assert.Equal(t, change.GetFieldNewValue(), float64(2))
			assert.Nil(t, change.GetFieldOldValue())
		}
		newSession.Close()
	}
}

func whatChanged_what_Changed_Array_Value_Removed(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		arr := &Arr{}
		arr.Array = []ravendb.Object{"a", 1, "b"}
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

		arr.Array = []ravendb.Object{"a"}

		changes, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 1)
		change := changes["arr/1"]
		assert.Equal(t, len(change), 2)

		{
			change := change[0]
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_ARRAY_VALUE_REMOVED)
			assert.Equal(t, change.GetFieldOldValue(), float64(1))
			assert.Nil(t, change.GetFieldNewValue())
		}

		{
			change := change[1]
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_ARRAY_VALUE_REMOVED)

			oldValStr := fmt.Sprintf("%#v", change.GetFieldOldValue())
			assert.Equal(t, oldValStr, "\"b\"")
			assert.Nil(t, change.GetFieldNewValue())
		}
		newSession.Close()
	}
}

func whatChanged_ravenDB_8169(t *testing.T) {
	//Test that when old and new values are of different type
	//but have the same value, we consider them unchanged

	var err error
	store := getDocumentStoreMust(t)
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

func whatChanged_whatChanged_should_be_idempotent_operation(t *testing.T) {
	//RavenDB-9150
	var err error
	store := getDocumentStoreMust(t)
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
	Array []ravendb.Object
}

func TestWhatChanged(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	whatChanged_what_Changed_Array_Value_Removed(t)
	whatChanged_whatChangedNewField(t)
	whatChanged_what_Changed_Array_Value_Added(t)
	whatChanged_whatChangedChangeField(t)
	whatChanged_whatChangedArrayValueChanged(t)
	whatChanged_ravenDB_8169(t)
	whatChanged_whatChangedRemovedField(t)
	whatChanged_whatChanged_should_be_idempotent_operation(t)
}
