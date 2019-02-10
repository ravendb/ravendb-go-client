package tests

import (
	"fmt"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

type Family struct {
	Names []string
}

type FamilyMembers struct {
	Members []*Member
}

type Member struct {
	Name string
	Age  int
}

type Arr1 struct {
	Str []string
}

type Arr2 struct {
	Arr1 []*Arr1
}

type Poc struct {
	Name string
	Obj  *User
}

func crudTestEntitiesAreSavedUsingLowerCase(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		user1 := &User{}
		user1.setLastName("user1")

		err = newSession.StoreWithID(user1, "users/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}

	documentsCommand, err := ravendb.NewGetDocumentsCommand([]string{"users/1"}, nil, false)
	assert.NoError(t, err)
	err = store.GetRequestExecutor("").ExecuteCommand(documentsCommand, nil)
	assert.NoError(t, err)

	result := documentsCommand.Result
	userJSON := result.Results[0]
	_, exists := userJSON["lastName"]
	assert.True(t, exists)

	{
		newSession := openSessionMust(t, store)
		var users []*User
		q := newSession.Advanced().RawQuery("from Users where lastName = 'user1'")
		err = q.GetResults(&users)
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)

		newSession.Close()
	}
}

func crudTestCanCustomizePropertyNamingStrategy(t *testing.T, driver *RavenTestDriver) {
	// Note: not possible to tweak behavior of JSON serialization
	// (entity mapper) in Go
}

func crudTestCrudOperations(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)

		user1 := &User{}
		user1.setLastName("user1")
		err = newSession.StoreWithID(user1, "users/1")
		assert.NoError(t, err)

		user2 := &User{}
		user2.setName("user2")
		user1.Age = 1
		err = newSession.StoreWithID(user2, "users/2")
		assert.NoError(t, err)

		user3 := &User{}
		user3.setName("user3")
		user3.Age = 1
		err = newSession.StoreWithID(user3, "users/3")
		assert.NoError(t, err)

		user4 := &User{}
		user4.setName("user4")
		err = newSession.StoreWithID(user4, "users/4")
		assert.NoError(t, err)

		err = newSession.DeleteEntity(user2)
		assert.NoError(t, err)
		user3.Age = 3
		err = newSession.SaveChanges()
		assert.NoError(t, err)

		var tempUser *User
		err = newSession.Load(&tempUser, "users/2")
		assert.NoError(t, err)
		assert.Nil(t, tempUser)

		tempUser = nil
		err = newSession.Load(&tempUser, "users/3")
		assert.NoError(t, err)
		assert.Equal(t, tempUser.Age, 3)

		user1 = nil
		err = newSession.Load(&user1, "users/1")
		assert.NoError(t, err)

		user4 = nil
		err = newSession.Load(&user4, "users/4")
		assert.NoError(t, err)

		err = newSession.DeleteEntity(user4)
		assert.NoError(t, err)
		user1.Age = 10
		err = newSession.SaveChanges()
		assert.NoError(t, err)

		tempUser = nil
		err = newSession.Load(&tempUser, "users/4")
		assert.NoError(t, err)
		assert.Nil(t, tempUser)
		tempUser = nil
		err = newSession.Load(&tempUser, "users/1")
		assert.NoError(t, err)
		assert.Equal(t, tempUser.Age, 10)
		newSession.Close()
	}
}

func crudTestCrudOperationsWithWhatChanged(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)

		{
			user1 := &User{}
			user1.setLastName("user1")
			err = newSession.StoreWithID(user1, "users/1")
			assert.NoError(t, err)

			user2 := &User{}
			user2.setName("user2")
			user1.Age = 1 // TODO: that's probably a bug in Java code
			err = newSession.StoreWithID(user2, "users/2")
			assert.NoError(t, err)

			user3 := &User{}
			user3.setName("user3")
			user3.Age = 1
			err = newSession.StoreWithID(user3, "users/3")
			assert.NoError(t, err)

			user4 := &User{}
			user4.setName("user4")
			err = newSession.StoreWithID(user4, "users/4")
			assert.NoError(t, err)

			err = newSession.DeleteEntity(user2)
			assert.NoError(t, err)
			user3.Age = 3

			changes, _ := newSession.Advanced().WhatChanged()
			assert.Equal(t, len(changes), 4)

			err = newSession.SaveChanges()
			assert.NoError(t, err)
		}

		{
			var user1, user2, user3, user4 *User
			err = newSession.Load(&user2, "users/2")
			assert.NoError(t, err)
			assert.Nil(t, user2)

			err = newSession.Load(&user3, "users/3")
			assert.NoError(t, err)
			assert.Equal(t, user3.Age, 3)

			err = newSession.Load(&user1, "users/1")
			assert.NoError(t, err)
			assert.NotNil(t, user1)

			err = newSession.Load(&user4, "users/4")
			assert.NoError(t, err)
			assert.NotNil(t, user4)

			err = newSession.DeleteEntity(user4)
			assert.NoError(t, err)

			user1.Age = 10

			var changes map[string][]*ravendb.DocumentsChanges
			changes, err = newSession.Advanced().WhatChanged()
			assert.NoError(t, err)
			assert.Equal(t, len(changes), 2)

			err = newSession.SaveChanges()
			assert.NoError(t, err)

		}

		var tempUser *User
		err = newSession.Load(&tempUser, "users/4")
		assert.NoError(t, err)
		assert.Nil(t, tempUser)

		tempUser = nil
		err = newSession.Load(&tempUser, "users/1")
		assert.NoError(t, err)
		assert.Equal(t, tempUser.Age, 10)
		newSession.Close()
	}
}

func crudTestCrudOperationsWithArrayInObject(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		family := &Family{}
		family.Names = []string{"Hibernating Rhinos", "RavenDB"}
		err = newSession.StoreWithID(family, "family/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)

		var newFamily *Family
		err = newSession.Load(&newFamily, "family/1")
		assert.NoError(t, err)
		newFamily.Names = []string{"Toli", "Mitzi", "Boki"}
		changes, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 1)
		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}
}

func crudTestCrudOperationsWithArrayInObject2(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		family := &Family{}
		family.Names = []string{"Hibernating Rhinos", "RavenDB"}
		err = newSession.StoreWithID(family, "family/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)

		var newFamily *Family
		err = newSession.Load(&newFamily, "family/1")
		assert.NoError(t, err)
		newFamily.Names = []string{"Hibernating Rhinos", "RavenDB"}
		changes, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 0)

		newFamily.Names = []string{"RavenDB", "Hibernating Rhinos"}
		changes, _ = newSession.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 1)

		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}
}

func crudTestCrudOperationsWithArrayInObject3(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		family := &Family{}
		family.Names = []string{"Hibernating Rhinos", "RavenDB"}
		err = newSession.StoreWithID(family, "family/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)

		var newFamily *Family
		err = newSession.Load(&newFamily, "family/1")
		assert.NoError(t, err)
		newFamily.Names = []string{"RavenDB"}
		changes, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 1)

		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}
}

func crudTestCrudOperationsWithArrayInObject4(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		family := &Family{}
		family.Names = []string{"Hibernating Rhinos", "RavenDB"}
		err = newSession.StoreWithID(family, "family/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)

		var newFamily *Family
		err = newSession.Load(&newFamily, "family/1")
		assert.NoError(t, err)
		newFamily.Names = []string{"RavenDB", "Hibernating Rhinos", "Toli", "Mitzi", "Boki"}
		changes, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 1)

		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}
}

func crudTestCrudOperationsWithNull(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		user := &User{}

		err = newSession.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)

		var user2 *User
		err = newSession.Load(&user2, "users/1")
		assert.NoError(t, err)
		WhatChanged, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(WhatChanged), 0)

		user2.Age = 3
		WhatChanged, _ = newSession.Advanced().WhatChanged()
		assert.Equal(t, len(WhatChanged), 1)
		newSession.Close()
	}
}

func crudTestCrudOperationsWithArrayOfObjects(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)

		member1 := &Member{}
		member1.Name = "Hibernating Rhinos"
		member1.Age = 8

		member2 := &Member{}
		member2.Name = "RavenDB"
		member2.Age = 4

		family := &FamilyMembers{}
		family.Members = []*Member{member1, member2}

		err = newSession.StoreWithID(family, "family/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)

		member1 = &Member{}
		member1.Name = "RavenDB"
		member1.Age = 4

		member2 = &Member{}
		member2.Name = "Hibernating Rhinos"
		member2.Age = 8

		var newFamily *FamilyMembers
		err = newSession.Load(&newFamily, "family/1")
		assert.NoError(t, err)
		newFamily.Members = []*Member{member1, member2}

		changes, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 1)

		family1Changes := changes["family/1"]
		assert.Equal(t, len(family1Changes), 4)

		// Note: order or fields differs from Java. In Java the order seems to be the order
		// of declaration in a class. In Go it's alphabetical
		{
			change := family1Changes[0]
			assert.Equal(t, change.FieldName, "Age")
			assert.Equal(t, change.Change, ravendb.DocumentChangeFieldChanged)
			oldVal := change.FieldOldValue
			assert.Equal(t, oldVal, 8.0)
			newVal := change.FieldNewValue
			assert.Equal(t, newVal, 4.0)
		}

		{
			change := family1Changes[1]
			assert.Equal(t, change.FieldName, "Name")
			assert.Equal(t, change.Change, ravendb.DocumentChangeFieldChanged)
			oldValStr := fmt.Sprintf("%#v", change.FieldOldValue)
			assert.Equal(t, oldValStr, "\"Hibernating Rhinos\"")
			newValStr := fmt.Sprintf("%#v", change.FieldNewValue)
			assert.Equal(t, newValStr, "\"RavenDB\"")
		}

		{
			change := family1Changes[2]
			assert.Equal(t, change.FieldName, "Age")
			assert.Equal(t, change.Change, ravendb.DocumentChangeFieldChanged)
			oldVal := change.FieldOldValue
			assert.Equal(t, oldVal, 4.0)
			newVal := change.FieldNewValue
			assert.Equal(t, newVal, 8.0)
		}

		{
			change := family1Changes[3]
			assert.Equal(t, change.FieldName, "Name")
			assert.Equal(t, change.Change, ravendb.DocumentChangeFieldChanged)
			oldValStr := fmt.Sprintf("%#v", change.FieldOldValue)
			assert.Equal(t, oldValStr, "\"RavenDB\"")
			newValStr := fmt.Sprintf("%#v", change.FieldNewValue)
			assert.Equal(t, newValStr, "\"Hibernating Rhinos\"")
		}

		member1 = &Member{}
		member1.Name = "Toli"
		member1.Age = 5

		member2 = &Member{}
		member2.Name = "Boki"
		member2.Age = 15

		newFamily.Members = []*Member{member1, member2}
		changes, _ = newSession.Advanced().WhatChanged()

		assert.Equal(t, len(changes), 1)

		family1Changes = changes["family/1"]
		assert.Equal(t, len(family1Changes), 4)

		// Note: the order of fields in Go is different than in Java. In Go it's alphabetic.
		{
			change := family1Changes[0]
			assert.Equal(t, change.FieldName, "Age")
			assert.Equal(t, change.Change, ravendb.DocumentChangeFieldChanged)
			oldVal := change.FieldOldValue
			assert.Equal(t, oldVal, 8.0)
			newVal := change.FieldNewValue
			assert.Equal(t, newVal, 5.0)
		}

		{
			change := family1Changes[1]
			assert.Equal(t, change.FieldName, "Name")
			assert.Equal(t, change.Change, ravendb.DocumentChangeFieldChanged)
			oldValStr := fmt.Sprintf("%#v", change.FieldOldValue)
			assert.Equal(t, oldValStr, "\"Hibernating Rhinos\"")
			newValStr := fmt.Sprintf("%#v", change.FieldNewValue)
			assert.Equal(t, newValStr, "\"Toli\"")
		}

		{
			change := family1Changes[2]
			assert.Equal(t, change.FieldName, "Age")
			assert.Equal(t, change.Change, ravendb.DocumentChangeFieldChanged)
			oldVal := change.FieldOldValue
			assert.Equal(t, oldVal, 4.0)
			newVal := change.FieldNewValue
			assert.Equal(t, newVal, 15.0)
		}

		{
			change := family1Changes[3]
			assert.Equal(t, change.FieldName, "Name")
			assert.Equal(t, change.Change, ravendb.DocumentChangeFieldChanged)
			oldValStr := fmt.Sprintf("%#v", change.FieldOldValue)
			assert.Equal(t, oldValStr, "\"RavenDB\"")
			newValStr := fmt.Sprintf("%#v", change.FieldNewValue)
			assert.Equal(t, newValStr, "\"Boki\"")
		}
		newSession.Close()
	}
}

func crudTestCrudOperationsWithArrayOfArrays(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		a1 := &Arr1{}
		a1.Str = []string{"a", "b"}

		a2 := &Arr1{}
		a2.Str = []string{"c", "d"}

		arr := &Arr2{}
		arr.Arr1 = []*Arr1{a1, a2}

		err = newSession.StoreWithID(arr, "arr/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)

		var newArr *Arr2
		err = newSession.Load(&newArr, "arr/1")
		assert.NoError(t, err)

		a1 = &Arr1{}
		a1.Str = []string{"d", "c"}

		a2 = &Arr1{}
		a2.Str = []string{"a", "b"}

		newArr.Arr1 = []*Arr1{a1, a2}

		WhatChanged, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, 1, len(WhatChanged))

		change := WhatChanged["arr/1"]
		assert.Equal(t, len(change), 4)

		{
			oldValueStr := fmt.Sprintf("%#v", change[0].FieldOldValue)
			assert.Equal(t, oldValueStr, "\"a\"")
			newValueStr := fmt.Sprintf("%#v", change[0].FieldNewValue)
			assert.Equal(t, newValueStr, "\"d\"")
		}

		{
			oldValueStr := fmt.Sprintf("%#v", change[1].FieldOldValue)
			assert.Equal(t, oldValueStr, "\"b\"")
			newValueStr := fmt.Sprintf("%#v", change[1].FieldNewValue)
			assert.Equal(t, newValueStr, "\"c\"")
		}

		{
			oldValueStr := fmt.Sprintf("%#v", change[2].FieldOldValue)
			assert.Equal(t, oldValueStr, "\"c\"")
			newValueStr := fmt.Sprintf("%#v", change[2].FieldNewValue)
			assert.Equal(t, newValueStr, "\"a\"")
		}

		{
			oldValueStr := fmt.Sprintf("%#v", change[3].FieldOldValue)
			assert.Equal(t, oldValueStr, "\"d\"")
			newValueStr := fmt.Sprintf("%#v", change[3].FieldNewValue)
			assert.Equal(t, newValueStr, "\"b\"")
		}

		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}

	{
		newSession := openSessionMust(t, store)
		var newArr *Arr2
		err = newSession.Load(&newArr, "arr/1")
		assert.NoError(t, err)
		a1 := &Arr1{}
		a1.Str = []string{"q", "w"}

		a2 := &Arr1{}
		a2.Str = []string{"a", "b"}
		newArr.Arr1 = []*Arr1{a1, a2}

		WhatChanged, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(WhatChanged), 1)

		change := WhatChanged["arr/1"]
		assert.Equal(t, len(change), 2)

		{
			oldValueStr := fmt.Sprintf("%#v", change[0].FieldOldValue)
			assert.Equal(t, oldValueStr, "\"d\"")
			newValueStr := fmt.Sprintf("%#v", change[0].FieldNewValue)
			assert.Equal(t, newValueStr, "\"q\"")
		}

		{
			oldValueStr := fmt.Sprintf("%#v", change[1].FieldOldValue)
			assert.Equal(t, oldValueStr, "\"c\"")
			newValueStr := fmt.Sprintf("%#v", change[1].FieldNewValue)
			assert.Equal(t, newValueStr, "\"w\"")
		}

		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}
}

func crudTestCrudCanUpdatePropertyToNull(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		{
			newSession := openSessionMust(t, store)
			user1 := &User{}
			user1.setLastName("user1")
			err = newSession.StoreWithID(user1, "users/1")
			assert.NoError(t, err)
			err = newSession.SaveChanges()
			assert.NoError(t, err)
			newSession.Close()
		}

		{
			newSession := openSessionMust(t, store)
			var user *User
			err = newSession.Load(&user, "users/1")
			assert.NoError(t, err)
			user.Name = nil
			err = newSession.SaveChanges()
			assert.NoError(t, err)
			newSession.Close()
		}

		{
			newSession := openSessionMust(t, store)
			var user *User
			err = newSession.Load(&user, "users/1")
			assert.NoError(t, err)
			assert.Nil(t, user.Name)
			newSession.Close()
		}
	}
}

func crudTestCrudCanUpdatePropertyFromNullToObject(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		poc := &Poc{}
		poc.Name = "aviv"

		err = session.StoreWithID(poc, "pocs/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		session := openSessionMust(t, store)
		var poc *Poc
		err = session.Load(&poc, "pocs/1")
		assert.NoError(t, err)
		assert.Nil(t, poc.Obj)

		user := &User{}
		poc.Obj = user
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		session := openSessionMust(t, store)
		var poc *Poc
		err = session.Load(&poc, "pocs/1")
		assert.NoError(t, err)
		assert.NotNil(t, poc.Obj)
		session.Close()
	}
}

func TestCrud(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	crudTestCrudOperationsWithNull(t, driver)
	crudTestCrudOperationsWithArrayOfObjects(t, driver)
	crudTestCrudOperationsWithWhatChanged(t, driver)
	crudTestCrudOperations(t, driver)
	crudTestCrudOperationsWithArrayInObject(t, driver)
	crudTestCrudCanUpdatePropertyToNull(t, driver)
	crudTestEntitiesAreSavedUsingLowerCase(t, driver)
	crudTestCanCustomizePropertyNamingStrategy(t, driver)
	crudTestCrudCanUpdatePropertyFromNullToObject(t, driver)
	crudTestCrudOperationsWithArrayInObject2(t, driver)
	crudTestCrudOperationsWithArrayInObject3(t, driver)
	crudTestCrudOperationsWithArrayInObject4(t, driver)
	crudTestCrudOperationsWithArrayOfArrays(t, driver)
}
