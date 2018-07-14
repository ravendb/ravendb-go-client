package ravendb

import (
	"fmt"
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

func whatChanged_whatChangedNewField(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)

	{
		newSession := openSessionMust(t, store)
		basicName := &BasicName{}
		basicName.setName("Toli")
		err = newSession.StoreEntityWithID(basicName, "users/1")
		assert.NoError(t, err)

		changes := newSession.advanced().whatChanged()
		assert.Equal(t, len(changes), 1)
		err = newSession.SaveChanges()
		assert.NoError(t, err)
	}

	{
		newSession := openSessionMust(t, store)
		userI, err := newSession.load(getTypeOf(&NameAndAge{}), "users/1")
		assert.NoError(t, err)
		user := userI.(*NameAndAge)
		user.setAge(5)

		changes := newSession.advanced().whatChanged()
		change := changes["users/1"]
		assert.Equal(t, len(change), 1)

		{
			change := change[0]
			assert.Equal(t, change.getChange(), DocumentsChanges_ChangeType_NEW_FIELD)
			err = newSession.SaveChanges()
			assert.NoError(t, err)
		}
	}
}

func whatChanged_whatChangedRemovedField(t *testing.T) {
}

func whatChanged_whatChangedChangeField(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)

	{
		newSession := openSessionMust(t, store)
		basicAge := &BasicAge{}
		basicAge.setAge(5)
		err = newSession.StoreEntityWithID(basicAge, "users/1")
		assert.NoError(t, err)

		changes := newSession.advanced().whatChanged()
		assert.Equal(t, len(changes), 1)

		err = newSession.SaveChanges()
		assert.NoError(t, err)
	}

	{
		newSession := openSessionMust(t, store)
		_, err = newSession.load(getTypeOf(&Int{}), "users/1")
		assert.NoError(t, err)
		changes := newSession.advanced().whatChanged()
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
	}
}

func whatChanged_whatChangedArrayValueChanged(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)

	{
		newSession := openSessionMust(t, store)
		arr := &Arr{}
		arr.setArray([]Object{"a", 1, "b"})

		err = newSession.StoreEntityWithID(arr, "users/1")
		assert.NoError(t, err)
		changes := newSession.advanced().whatChanged()
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
			arrI, err := newSession.load(getTypeOf(&Arr{}), "users/1")
			assert.NoError(t, err)
			arr := arrI.(*Arr)

			arr.setArray([]Object{"a", 2, "c"})

			changes := newSession.advanced().whatChanged()
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
		}
	}
}

func whatChanged_what_Changed_Array_Value_Added(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)

	{
		newSession := openSessionMust(t, store)
		arr := &Arr{}
		arr.setArray([]Object{"a", 1, "b"})
		err = newSession.StoreEntityWithID(arr, "arr/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)
	}

	{
		newSession := openSessionMust(t, store)
		arrI, err := newSession.load(getTypeOf(&Arr{}), "arr/1")
		assert.NoError(t, err)

		arr := arrI.(*Arr)
		arr.setArray([]Object{"a", 1, "b", "c", 2})

		changes := newSession.advanced().whatChanged()
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
	}
}

func whatChanged_what_Changed_Array_Value_Removed(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)

	{
		newSession := openSessionMust(t, store)
		arr := &Arr{}
		arr.setArray([]Object{"a", 1, "b"})
		err = newSession.StoreEntityWithID(arr, "arr/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)
	}

	{
		newSession := openSessionMust(t, store)
		arrI, err := newSession.load(getTypeOf(&Arr{}), "arr/1")
		assert.NoError(t, err)

		arr := arrI.(*Arr)
		arr.setArray([]Object{"a"})

		changes := newSession.advanced().whatChanged()
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

	}
}

func whatChanged_ravenDB_8169(t *testing.T) {
	//Test that when old and new values are of different type
	//but have the same value, we consider them unchanged

	var err error
	store := getDocumentStoreMust(t)

	{
		newSession := openSessionMust(t, store)

		anInt := &Int{}
		anInt.setNumber(1)

		err = newSession.StoreEntityWithID(anInt, "num/1")
		assert.NoError(t, err)

		aDouble := &Double{}
		aDouble.setNumber(2.0)
		err = newSession.StoreEntityWithID(aDouble, "num/2")
		assert.NoError(t, err)

		err = newSession.SaveChanges()
		assert.NoError(t, err)
	}

	{
		newSession := openSessionMust(t, store)
		_, err = newSession.load(getTypeOf(&Double{}), "num/1")
		assert.NoError(t, err)

		changes := newSession.advanced().whatChanged()
		assert.Equal(t, len(changes), 0)
	}

	{
		newSession := openSessionMust(t, store)
		_, err = newSession.load(getTypeOf(&Int{}), "num/2")
		assert.NoError(t, err)

		changes := newSession.advanced().whatChanged()
		assert.Equal(t, len(changes), 0)
	}
}

func whatChanged_whatChanged_should_be_idempotent_operation(t *testing.T) {
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
	if useProxy() {
		proxy.ChangeLogFile("trace_what_changed_go.txt")
	}

	createTestDriver()
	defer deleteTestDriver()

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
