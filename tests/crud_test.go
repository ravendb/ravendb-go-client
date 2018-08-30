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

func (f *Family) GetNames() []string {
	return f.Names
}

func (f *Family) setNames(names []string) {
	f.Names = names
}

type FamilyMembers struct {
	Members []*Member
}

func (m *FamilyMembers) getMembers() []*Member {
	return m.Members
}

func (m *FamilyMembers) setMembers(members []*Member) {
	m.Members = members
}

type Member struct {
	Name string
	Age  int
}

func (m *Member) GetName() string {
	return m.Name
}

func (m *Member) setName(name string) {
	m.Name = name
}

func (m *Member) getAge() int {
	return m.Age
}

func (m *Member) setAge(age int) {
	m.Age = age
}

type Arr1 struct {
	Str []string
}

func (a *Arr1) getStr() []string {
	return a.Str
}

func (a *Arr1) setStr(str []string) {
	a.Str = str
}

type Arr2 struct {
	Arr1 []*Arr1
}

func (a *Arr2) getArr1() []*Arr1 {
	return a.Arr1
}

func (a *Arr2) setArr1(arr1 []*Arr1) {
	a.Arr1 = arr1
}

type Poc struct {
	Name string
	Obj  *User
}

func (p *Poc) GetName() string {
	return p.Name
}

func (p *Poc) setName(name string) {
	p.Name = name
}

func (p *Poc) getObj() *User {
	return p.Obj
}

func (p *Poc) setObj(obj *User) {
	p.Obj = obj
}

func crudTest_entitiesAreSavedUsingLowerCase(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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

	documentsCommand := ravendb.NewGetDocumentsCommand([]string{"users/1"}, nil, false)
	err = store.GetRequestExecutor().ExecuteCommand(documentsCommand)
	assert.NoError(t, err)

	result := documentsCommand.Result
	userJson := result.Results[0]
	_, exists := userJson["lastName"]
	assert.True(t, exists)

	{
		newSession := openSessionMust(t, store)
		users, err := newSession.Advanced().RawQuery(ravendb.GetTypeOf(&User{}), "from Users where lastName = 'user1'").ToList()
		assert.NoError(t, err)

		assert.Equal(t, len(users), 1)

		newSession.Close()
	}
}

func crudTest_canCustomizePropertyNamingStrategy(t *testing.T) {
	// Note: not possible to tweak behavior of JSON serialization
	// (entity mapper) in Go
}

func crudTest_crudOperations(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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

		tempUserI, err := newSession.LoadOld(ravendb.GetTypeOf(&User{}), "users/2")
		assert.NoError(t, err)
		tempUser := tempUserI.(*User)
		assert.Nil(t, tempUser)

		tempUserI, err = newSession.LoadOld(ravendb.GetTypeOf(&User{}), "users/3")
		assert.NoError(t, err)
		tempUser = tempUserI.(*User)
		assert.Equal(t, tempUser.Age, 3)

		user1I, err := newSession.LoadOld(ravendb.GetTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		user1 = user1I.(*User)
		user4I, err := newSession.LoadOld(ravendb.GetTypeOf(&User{}), "users/4")
		assert.NoError(t, err)
		user4 = user4I.(*User)

		err = newSession.DeleteEntity(user4)
		assert.NoError(t, err)
		user1.Age = 10
		err = newSession.SaveChanges()
		assert.NoError(t, err)

		tempUserI, err = newSession.LoadOld(ravendb.GetTypeOf(&User{}), "users/4")
		tempUser = tempUserI.(*User)
		assert.Nil(t, tempUser)
		tempUserI, err = newSession.LoadOld(ravendb.GetTypeOf(&User{}), "users/1")
		tempUser = tempUserI.(*User)
		assert.Equal(t, tempUser.Age, 10)
		newSession.Close()
	}
}

func crudTest_crudOperationsWithWhatChanged(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)

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

		tempUserI, err := newSession.LoadOld(ravendb.GetTypeOf(&User{}), "users/2")
		assert.NoError(t, err)
		tempUser := tempUserI.(*User)
		assert.Nil(t, tempUser)

		tempUserI, err = newSession.LoadOld(ravendb.GetTypeOf(&User{}), "users/3")
		assert.NoError(t, err)
		tempUser = tempUserI.(*User)
		assert.Equal(t, tempUser.Age, 3)

		user1I, err := newSession.LoadOld(ravendb.GetTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		user1 = user1I.(*User)

		user4I, err := newSession.LoadOld(ravendb.GetTypeOf(&User{}), "users/4")
		assert.NoError(t, err)
		user4 = user4I.(*User)

		err = newSession.DeleteEntity(user4)
		assert.NoError(t, err)

		user1.Age = 10

		if ravendb.EnableFailingTests {
			// TODO: this returns 3 changes, showing user/2 as added
			// which is probably wrong. Need to figure out why.
			changes, err := newSession.Advanced().WhatChanged()
			assert.NoError(t, err)
			assert.Equal(t, len(changes), 2)
		}

		err = newSession.SaveChanges()
		assert.NoError(t, err)

		tempUserI, err = newSession.LoadOld(ravendb.GetTypeOf(&User{}), "users/4")
		assert.NoError(t, err)
		tempUser = tempUserI.(*User)
		assert.Nil(t, tempUser)

		tempUserI, err = newSession.LoadOld(ravendb.GetTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		tempUser = tempUserI.(*User)
		assert.Equal(t, tempUser.Age, 10)
		newSession.Close()
	}
}

func crudTest_crudOperationsWithArrayInObject(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		family := &Family{}
		family.setNames([]string{"Hibernating Rhinos", "RavenDB"})
		err = newSession.StoreWithID(family, "family/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)

		newFamilyI, err := newSession.LoadOld(ravendb.GetTypeOf(&Family{}), "family/1")
		assert.NoError(t, err)
		newFamily := newFamilyI.(*Family)
		newFamily.setNames([]string{"Toli", "Mitzi", "Boki"})
		changes, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 1)
		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}
}

func crudTest_crudOperationsWithArrayInObject2(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		family := &Family{}
		family.setNames([]string{"Hibernating Rhinos", "RavenDB"})
		err = newSession.StoreWithID(family, "family/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)

		newFamilyI, err := newSession.LoadOld(ravendb.GetTypeOf(&Family{}), "family/1")
		assert.NoError(t, err)
		newFamily := newFamilyI.(*Family)
		newFamily.setNames([]string{"Hibernating Rhinos", "RavenDB"})
		changes, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 0)

		newFamily.setNames([]string{"RavenDB", "Hibernating Rhinos"})
		changes, _ = newSession.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 1)

		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}
}

func crudTest_crudOperationsWithArrayInObject3(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		family := &Family{}
		family.setNames([]string{"Hibernating Rhinos", "RavenDB"})
		err = newSession.StoreWithID(family, "family/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)

		newFamilyI, err := newSession.LoadOld(ravendb.GetTypeOf(&Family{}), "family/1")
		assert.NoError(t, err)
		newFamily := newFamilyI.(*Family)
		newFamily.setNames([]string{"RavenDB"})
		changes, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 1)

		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}
}

func crudTest_crudOperationsWithArrayInObject4(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		family := &Family{}
		family.setNames([]string{"Hibernating Rhinos", "RavenDB"})
		err = newSession.StoreWithID(family, "family/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)

		newFamilyI, err := newSession.LoadOld(ravendb.GetTypeOf(&Family{}), "family/1")
		assert.NoError(t, err)
		newFamily := newFamilyI.(*Family)
		newFamily.setNames([]string{"RavenDB", "Hibernating Rhinos", "Toli", "Mitzi", "Boki"})
		changes, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 1)

		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}
}

func crudTest_crudOperationsWithNull(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		user := &User{}

		err = newSession.StoreWithID(user, "users/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)

		user2I, err := newSession.LoadOld(ravendb.GetTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		WhatChanged, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(WhatChanged), 0)

		user2 := user2I.(*User)
		user2.Age = 3
		WhatChanged, _ = newSession.Advanced().WhatChanged()
		assert.Equal(t, len(WhatChanged), 1)
		newSession.Close()
	}
}

func crudTest_crudOperationsWithArrayOfObjects(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)

		member1 := &Member{}
		member1.setName("Hibernating Rhinos")
		member1.Age = 8

		member2 := &Member{}
		member2.setName("RavenDB")
		member2.Age = 4

		family := &FamilyMembers{}
		family.setMembers([]*Member{member1, member2})

		err = newSession.StoreWithID(family, "family/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)

		member1 = &Member{}
		member1.setName("RavenDB")
		member1.Age = 4

		member2 = &Member{}
		member2.setName("Hibernating Rhinos")
		member2.Age = 8

		newFamilyI, err := newSession.LoadOld(ravendb.GetTypeOf(&FamilyMembers{}), "family/1")
		assert.NoError(t, err)
		newFamily := newFamilyI.(*FamilyMembers)
		newFamily.setMembers([]*Member{member1, member2})

		changes, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(changes), 1)

		family1Changes := changes["family/1"]
		assert.Equal(t, len(family1Changes), 4)

		// Note: order or fields differs from Java. In Java the order seems to be the order
		// of declaration in a class. In Go it's alphabetical
		{
			change := family1Changes[0]
			assert.Equal(t, change.GetFieldName(), "Age")
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_FIELD_CHANGED)
			oldValStr := fmt.Sprintf("%#v", change.GetFieldOldValue())
			assert.Equal(t, oldValStr, "8")
			newValStr := fmt.Sprintf("%#v", change.GetFieldNewValue())
			assert.Equal(t, newValStr, "4")
		}

		{
			change := family1Changes[1]
			assert.Equal(t, change.GetFieldName(), "Name")
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_FIELD_CHANGED)
			oldValStr := fmt.Sprintf("%#v", change.GetFieldOldValue())
			assert.Equal(t, oldValStr, "\"Hibernating Rhinos\"")
			newValStr := fmt.Sprintf("%#v", change.GetFieldNewValue())
			assert.Equal(t, newValStr, "\"RavenDB\"")
		}

		{
			change := family1Changes[2]
			assert.Equal(t, change.GetFieldName(), "Age")
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_FIELD_CHANGED)
			oldValStr := fmt.Sprintf("%#v", change.GetFieldOldValue())
			assert.Equal(t, oldValStr, "4")
			newValStr := fmt.Sprintf("%#v", change.GetFieldNewValue())
			assert.Equal(t, newValStr, "8")
		}

		{
			change := family1Changes[3]
			assert.Equal(t, change.GetFieldName(), "Name")
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_FIELD_CHANGED)
			oldValStr := fmt.Sprintf("%#v", change.GetFieldOldValue())
			assert.Equal(t, oldValStr, "\"RavenDB\"")
			newValStr := fmt.Sprintf("%#v", change.GetFieldNewValue())
			assert.Equal(t, newValStr, "\"Hibernating Rhinos\"")
		}

		member1 = &Member{}
		member1.setName("Toli")
		member1.Age = 5

		member2 = &Member{}
		member2.setName("Boki")
		member2.Age = 15

		newFamily.setMembers([]*Member{member1, member2})
		changes, _ = newSession.Advanced().WhatChanged()

		assert.Equal(t, len(changes), 1)

		family1Changes = changes["family/1"]
		assert.Equal(t, len(family1Changes), 4)

		// Note: the order of fields in Go is different than in Java. In Go it's alphabetic.
		{
			change := family1Changes[0]
			assert.Equal(t, change.GetFieldName(), "Age")
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_FIELD_CHANGED)
			oldValStr := fmt.Sprintf("%#v", change.GetFieldOldValue())
			assert.Equal(t, oldValStr, "8")
			newValStr := fmt.Sprintf("%#v", change.GetFieldNewValue())
			assert.Equal(t, newValStr, "5")
		}

		{
			change := family1Changes[1]
			assert.Equal(t, change.GetFieldName(), "Name")
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_FIELD_CHANGED)
			oldValStr := fmt.Sprintf("%#v", change.GetFieldOldValue())
			assert.Equal(t, oldValStr, "\"Hibernating Rhinos\"")
			newValStr := fmt.Sprintf("%#v", change.GetFieldNewValue())
			assert.Equal(t, newValStr, "\"Toli\"")
		}

		{
			change := family1Changes[2]
			assert.Equal(t, change.GetFieldName(), "Age")
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_FIELD_CHANGED)
			oldValStr := fmt.Sprintf("%#v", change.GetFieldOldValue())
			assert.Equal(t, oldValStr, "4")
			newValStr := fmt.Sprintf("%#v", change.GetFieldNewValue())
			assert.Equal(t, newValStr, "15")
		}

		{
			change := family1Changes[3]
			assert.Equal(t, change.GetFieldName(), "Name")
			assert.Equal(t, change.GetChange(), ravendb.DocumentsChanges_ChangeType_FIELD_CHANGED)
			oldValStr := fmt.Sprintf("%#v", change.GetFieldOldValue())
			assert.Equal(t, oldValStr, "\"RavenDB\"")
			newValStr := fmt.Sprintf("%#v", change.GetFieldNewValue())
			assert.Equal(t, newValStr, "\"Boki\"")
		}
		newSession.Close()
	}
}

func crudTest_crudOperationsWithArrayOfArrays(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		a1 := &Arr1{}
		a1.setStr([]string{"a", "b"})

		a2 := &Arr1{}
		a2.setStr([]string{"c", "d"})

		arr := &Arr2{}
		arr.setArr1([]*Arr1{a1, a2})

		newSession.StoreWithID(arr, "arr/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)

		newArrI, err := newSession.LoadOld(ravendb.GetTypeOf(&Arr2{}), "arr/1")
		assert.NoError(t, err)
		newArr := newArrI.(*Arr2)

		a1 = &Arr1{}
		a1.setStr([]string{"d", "c"})

		a2 = &Arr1{}
		a2.setStr([]string{"a", "b"})

		newArr.setArr1([]*Arr1{a1, a2})

		WhatChanged, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, 1, len(WhatChanged))

		change := WhatChanged["arr/1"]
		assert.Equal(t, len(change), 4)

		{
			oldValueStr := fmt.Sprintf("%#v", change[0].GetFieldOldValue())
			assert.Equal(t, oldValueStr, "\"a\"")
			newValueStr := fmt.Sprintf("%#v", change[0].GetFieldNewValue())
			assert.Equal(t, newValueStr, "\"d\"")
		}

		{
			oldValueStr := fmt.Sprintf("%#v", change[1].GetFieldOldValue())
			assert.Equal(t, oldValueStr, "\"b\"")
			newValueStr := fmt.Sprintf("%#v", change[1].GetFieldNewValue())
			assert.Equal(t, newValueStr, "\"c\"")
		}

		{
			oldValueStr := fmt.Sprintf("%#v", change[2].GetFieldOldValue())
			assert.Equal(t, oldValueStr, "\"c\"")
			newValueStr := fmt.Sprintf("%#v", change[2].GetFieldNewValue())
			assert.Equal(t, newValueStr, "\"a\"")
		}

		{
			oldValueStr := fmt.Sprintf("%#v", change[3].GetFieldOldValue())
			assert.Equal(t, oldValueStr, "\"d\"")
			newValueStr := fmt.Sprintf("%#v", change[3].GetFieldNewValue())
			assert.Equal(t, newValueStr, "\"b\"")
		}

		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}

	{
		newSession := openSessionMust(t, store)
		newArrI, err := newSession.LoadOld(ravendb.GetTypeOf(&Arr2{}), "arr/1")
		assert.NoError(t, err)
		newArr := newArrI.(*Arr2)
		a1 := &Arr1{}
		a1.setStr([]string{"q", "w"})

		a2 := &Arr1{}
		a2.setStr([]string{"a", "b"})
		newArr.setArr1([]*Arr1{a1, a2})

		WhatChanged, _ := newSession.Advanced().WhatChanged()
		assert.Equal(t, len(WhatChanged), 1)

		change := WhatChanged["arr/1"]
		assert.Equal(t, len(change), 2)

		{
			oldValueStr := fmt.Sprintf("%#v", change[0].GetFieldOldValue())
			assert.Equal(t, oldValueStr, "\"d\"")
			newValueStr := fmt.Sprintf("%#v", change[0].GetFieldNewValue())
			assert.Equal(t, newValueStr, "\"q\"")
		}

		{
			oldValueStr := fmt.Sprintf("%#v", change[1].GetFieldOldValue())
			assert.Equal(t, oldValueStr, "\"c\"")
			newValueStr := fmt.Sprintf("%#v", change[1].GetFieldNewValue())
			assert.Equal(t, newValueStr, "\"w\"")
		}

		err = newSession.SaveChanges()
		assert.NoError(t, err)
		newSession.Close()
	}
}

func crudTest_crudCanUpdatePropertyToNull(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
			userI, err := newSession.LoadOld(ravendb.GetTypeOf(&User{}), "users/1")
			assert.NoError(t, err)
			user := userI.(*User)
			user.Name = nil
			err = newSession.SaveChanges()
			assert.NoError(t, err)
			newSession.Close()
		}

		{
			newSession := openSessionMust(t, store)
			userI, err := newSession.LoadOld(ravendb.GetTypeOf(&User{}), "users/1")
			assert.NoError(t, err)
			user := userI.(*User)
			assert.Nil(t, user.Name)
			newSession.Close()
		}
	}
}

func crudTest_crudCanUpdatePropertyFromNullToObject(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		poc := &Poc{}
		poc.setName("aviv")

		err = session.StoreWithID(poc, "pocs/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		session := openSessionMust(t, store)
		pocI, err := session.LoadOld(ravendb.GetTypeOf(&Poc{}), "pocs/1")
		assert.NoError(t, err)
		poc := pocI.(*Poc)
		assert.Nil(t, poc.getObj())

		user := &User{}
		poc.setObj(user)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		session := openSessionMust(t, store)
		pocI, err := session.LoadOld(ravendb.GetTypeOf(&Poc{}), "pocs/1")
		assert.NoError(t, err)
		poc := pocI.(*Poc)
		assert.NotNil(t, poc.getObj())
		session.Close()
	}
}

func TestCrud(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	crudTest_crudOperationsWithNull(t)
	crudTest_crudOperationsWithArrayOfObjects(t)
	crudTest_crudOperationsWithWhatChanged(t)
	crudTest_crudOperations(t)
	crudTest_crudOperationsWithArrayInObject(t)
	crudTest_crudCanUpdatePropertyToNull(t)
	crudTest_entitiesAreSavedUsingLowerCase(t)
	crudTest_canCustomizePropertyNamingStrategy(t)
	crudTest_crudCanUpdatePropertyFromNullToObject(t)
	crudTest_crudOperationsWithArrayInObject2(t)
	crudTest_crudOperationsWithArrayInObject3(t)
	crudTest_crudOperationsWithArrayInObject4(t)
	crudTest_crudOperationsWithArrayOfArrays(t)
}
