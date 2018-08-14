package ravendb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func whatChanged_whatChangedNewField(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		basicName := &BasicName{}
		basicName.setName("Toli")
		err = newSession.StoreWithID(basicName, "users/1")
		assert.NoError(t, err)

		changes, _ := newSession.Advanced().whatChanged()
		assert.Equal(t, len(changes), 1)
		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}

	{
		newSession := openSessionMust(t, store)
		userI, err := newSession.Load(getTypeOf(&NameAndAge{}), "users/1")
		assert.NoError(t, err)
		user := userI.(*NameAndAge)
		user.setAge(5)

		changes, _ := newSession.Advanced().whatChanged()
		change := changes["users/1"]
		assert.Equal(t, len(change), 1)

		{
			change := change[0]
			assert.Equal(t, change.getChange(), DocumentsChanges_ChangeType_NEW_FIELD)
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
		nameAndAge.setAge(5)
		nameAndAge.setName("Toli")

		err = newSession.StoreWithID(nameAndAge, "users/1")
		assert.NoError(t, err)

		changes, _ := newSession.Advanced().whatChanged()
		assert.Equal(t, len(changes), 1)

		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}

	{
		newSession := openSessionMust(t, store)
		_, err = newSession.Load(getTypeOf(&BasicAge{}), "users/1")
		assert.NoError(t, err)

		changes, _ := newSession.Advanced().whatChanged()
		change := changes["users/1"]
		assert.Equal(t, len(change), 1)

		{
			change := change[0]
			assert.Equal(t, change.getChange(), DocumentsChanges_ChangeType_REMOVED_FIELD)
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
		basicAge.setAge(5)
		err = newSession.StoreWithID(basicAge, "users/1")
		assert.NoError(t, err)

		changes, _ := newSession.Advanced().whatChanged()
		assert.Equal(t, len(changes), 1)

		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}

	{
		newSession := openSessionMust(t, store)
		_, err = newSession.Load(getTypeOf(&Int{}), "users/1")
		assert.NoError(t, err)
		changes, _ := newSession.Advanced().whatChanged()
		change := changes["users/1"]
		assert.Equal(t, len(change), 2)

		{
			change := change[0]
			assert.Equal(t, change.getChange(), DocumentsChanges_ChangeType_REMOVED_FIELD)
		}

		{
			change := change[1]
			assert.Equal(t, change.getChange(), DocumentsChanges_ChangeType_NEW_FIELD)
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
		arr.setArray([]Object{"a", 1, "b"})

		err = newSession.StoreWithID(arr, "users/1")
		assert.NoError(t, err)
		changes, _ := newSession.Advanced().whatChanged()
		assert.Equal(t, len(changes), 1)

		change := changes["users/1"]
		assert.Equal(t, len(change), 1)

		{
			change := change[0]
			assert.Equal(t, change.getChange(), DocumentsChanges_ChangeType_DOCUMENT_ADDED)
			err = newSession.SaveChanges()
			assert.NoError(t, err)
		}

		{
			newSession := openSessionMust(t, store)
			arrI, err := newSession.Load(getTypeOf(&Arr{}), "users/1")
			assert.NoError(t, err)
			arr := arrI.(*Arr)

			arr.setArray([]Object{"a", 2, "c"})

			changes, _ := newSession.Advanced().whatChanged()
			assert.Equal(t, len(changes), 1)

			change := changes["users/1"]
			assert.Equal(t, len(change), 2)

			{
				change := change[0]
				assert.Equal(t, change.getChange(), DocumentsChanges_ChangeType_ARRAY_VALUE_CHANGED)
				oldValueStr := fmt.Sprintf("%#v", change.getFieldOldValue())
				assert.Equal(t, oldValueStr, "1")
				newValue := change.getFieldNewValue()
				assert.Equal(t, newValue, float64(2))
			}

			{
				change := change[1]
				assert.Equal(t, change.getChange(), DocumentsChanges_ChangeType_ARRAY_VALUE_CHANGED)
				oldValueStr := fmt.Sprintf("%#v", change.getFieldOldValue())
				assert.Equal(t, oldValueStr, "\"b\"")
				newValueStr := fmt.Sprintf("%#v", change.getFieldNewValue())
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
		arr.setArray([]Object{"a", 1, "b"})
		err = newSession.StoreWithID(arr, "arr/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}

	{
		newSession := openSessionMust(t, store)
		arrI, err := newSession.Load(getTypeOf(&Arr{}), "arr/1")
		assert.NoError(t, err)

		arr := arrI.(*Arr)
		arr.setArray([]Object{"a", 1, "b", "c", 2})

		changes, _ := newSession.Advanced().whatChanged()
		assert.Equal(t, len(changes), 1)
		change := changes["arr/1"]
		assert.Equal(t, len(change), 2)

		{
			change := change[0]
			assert.Equal(t, change.getChange(), DocumentsChanges_ChangeType_ARRAY_VALUE_ADDED)
			newValStr := fmt.Sprintf("%#v", change.getFieldNewValue())
			assert.Equal(t, newValStr, "\"c\"")
			assert.Nil(t, change.getFieldOldValue())
		}
		{
			change := change[1]
			assert.Equal(t, change.getChange(), DocumentsChanges_ChangeType_ARRAY_VALUE_ADDED)
			assert.Equal(t, change.getFieldNewValue(), float64(2))
			assert.Nil(t, change.getFieldOldValue())
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
		arr.setArray([]Object{"a", 1, "b"})
		err = newSession.StoreWithID(arr, "arr/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}

	{
		newSession := openSessionMust(t, store)
		arrI, err := newSession.Load(getTypeOf(&Arr{}), "arr/1")
		assert.NoError(t, err)

		arr := arrI.(*Arr)
		arr.setArray([]Object{"a"})

		changes, _ := newSession.Advanced().whatChanged()
		assert.Equal(t, len(changes), 1)
		change := changes["arr/1"]
		assert.Equal(t, len(change), 2)

		{
			change := change[0]
			assert.Equal(t, change.getChange(), DocumentsChanges_ChangeType_ARRAY_VALUE_REMOVED)
			assert.Equal(t, change.getFieldOldValue(), float64(1))
			assert.Nil(t, change.getFieldNewValue())
		}

		{
			change := change[1]
			assert.Equal(t, change.getChange(), DocumentsChanges_ChangeType_ARRAY_VALUE_REMOVED)

			oldValStr := fmt.Sprintf("%#v", change.getFieldOldValue())
			assert.Equal(t, oldValStr, "\"b\"")
			assert.Nil(t, change.getFieldNewValue())
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
		anInt.setNumber(1)

		err = newSession.StoreWithID(anInt, "num/1")
		assert.NoError(t, err)

		aDouble := &Double{}
		aDouble.setNumber(2.0)
		err = newSession.StoreWithID(aDouble, "num/2")
		assert.NoError(t, err)

		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}

	{
		newSession := openSessionMust(t, store)
		_, err = newSession.Load(getTypeOf(&Double{}), "num/1")
		assert.NoError(t, err)

		changes, _ := newSession.Advanced().whatChanged()
		assert.Equal(t, len(changes), 0)
		newSession.Close()
	}

	{
		newSession := openSessionMust(t, store)
		_, err = newSession.Load(getTypeOf(&Int{}), "num/2")
		assert.NoError(t, err)

		changes, _ := newSession.Advanced().whatChanged()
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

		user1 := NewUser()
		user1.setName("user1")

		user2 := NewUser()
		user2.setName("user2")
		user2.setAge(1)

		user3 := NewUser()
		user3.setName("user3")
		user3.setAge(1)

		err = session.StoreWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.StoreWithID(user2, "users/2")
		assert.NoError(t, err)
		err = session.StoreWithID(user3, "users/3")
		assert.NoError(t, err)

		changes, _ := session.Advanced().whatChanged()
		assert.Equal(t, len(changes), 3)

		err = session.SaveChanges()
		assert.NoError(t, err)

		user1I, err := session.Load(getTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		user1 = user1I.(*User)

		user2I, err := session.Load(getTypeOf(&User{}), "users/2")
		assert.NoError(t, err)
		user2 = user2I.(*User)

		user1.setAge(10)
		err = session.DeleteEntity(user2)

		changes, _ = session.Advanced().whatChanged()
		assert.Equal(t, len(changes), 2)
		changes, _ = session.Advanced().whatChanged()
		assert.Equal(t, len(changes), 2)
		session.Close()
	}
}

type BasicName struct {
	Name string
}

func (n *BasicName) getName() string {
	return n.Name
}

func (n *BasicName) setName(Name string) {
	n.Name = Name
}

type NameAndAge struct {
	Name string
	Age  int
}

func (n *NameAndAge) getName() string {
	return n.Name
}

func (n *NameAndAge) setName(Name string) {
	n.Name = Name
}

func (n *NameAndAge) getAge() int {
	return n.Age
}

func (n *NameAndAge) setAge(Age int) {
	n.Age = Age
}

type BasicAge struct {
	Age int
}

func (a *BasicAge) getAge() int {
	return a.Age
}

func (a *BasicAge) setAge(Age int) {
	a.Age = Age
}

type Int struct {
	Number int
}

func (i *Int) getNumber() int {
	return i.Number
}

func (i *Int) setNumber(Number int) {
	i.Number = Number
}

type Double struct {
	Number float64
}

func (d *Double) getNumber() float64 {
	return d.Number
}

func (d *Double) setNumber(Number float64) {
	d.Number = Number
}

type Arr struct {
	Array []Object
}

func (a *Arr) getArray() []Object {
	return a.Array
}

func (a *Arr) setArray(Array []Object) {
	a.Array = Array
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
