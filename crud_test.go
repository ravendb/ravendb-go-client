package ravendb

import (
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

type Family struct {
	Names []string
}

func (f *Family) getNames() []string {
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

func (m *Member) getName() string {
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

func (p *Poc) getName() string {
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
}

func crudTest_canCustomizePropertyNamingStrategy(t *testing.T) {
}
func crudTest_crudOperations(t *testing.T) {
}
func crudTest_crudOperationsWithWhatChanged(t *testing.T) {
}
func crudTest_crudOperationsWithArrayInObject(t *testing.T) {
}
func crudTest_crudOperationsWithArrayInObject2(t *testing.T) {
}
func crudTest_crudOperationsWithArrayInObject3(t *testing.T) {
}
func crudTest_crudOperationsWithArrayInObject4(t *testing.T) {
}
func crudTest_crudOperationsWithNull(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	{
		newSession := openSessionMust(t, store)
		user := NewUser()

		err = newSession.StoreEntityWithID(user, "users/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)

		user2I, err := newSession.load(getTypeOf(&User{}), "users/1")
		assert.NoError(t, err)
		whatChanged := newSession.advanced().whatChanged()
		assert.Equal(t, len(whatChanged), 0)

		user2 := user2I.(*User)
		user2.setAge(3)
		whatChanged = newSession.advanced().whatChanged()
		assert.Equal(t, len(whatChanged), 1)
	}
}

func crudTest_crudOperationsWithArrayOfObjects(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	{
		newSession := openSessionMust(t, store)

		member1 := &Member{}
		member1.setName("Hibernating Rhinos")
		member1.setAge(8)

		member2 := &Member{}
		member2.setName("RavenDB")
		member2.setAge(4)

		family := &FamilyMembers{}
		family.setMembers([]*Member{member1, member2})

		err = newSession.StoreEntityWithID(family, "family/1")
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)

		member1 = &Member{}
		member1.setName("RavenDB")
		member1.setAge(4)

		member2 = &Member{}
		member2.setName("Hibernating Rhinos")
		member2.setAge(8)

		newFamilyI, err := newSession.load(getTypeOf(&FamilyMembers{}), "family/1")
		assert.NoError(t, err)
		newFamily := newFamilyI.(*FamilyMembers)
		newFamily.setMembers([]*Member{member1, member2})

		changes := newSession.advanced().whatChanged()
		assert.Equal(t, len(changes), 1)

		family1Changes := changes["family/1"]
		assert.Equal(t, len(family1Changes), 4)

		// TODO: Go randomizes traversal order of maps so changes for each object can be either Name/Age or Age/Map
		// we can fix it either by being smarter about how we test result or fixing the order in json_operation.go
		// e.g. by makeing getObjectNodeFieldNames() sort results
		/*
			{
				change := family1Changes[0]
				assert.Equal(t, change.getFieldName(), "Name")
				assert.Equal(t, change.getChange(), DocumentsChanges_ChangeType_FIELD_CHANGED)
				oldVal := change.getFieldOldValue()
				oldValStr := fmt.Sprintf("%s", oldVal)
				assert.Equal(t, oldValStr, "Hibernating Rhinos")

				newVal := change.getFieldNewValue()
				newValStr := fmt.Sprintf("%s", newVal)
				assert.Equal(t, newValStr, "RavenDB")
			}

			{
				change := family1Changes[1]
				assert.Equal(t, change1.getFieldName(), "Age")

				assert.Equal(t, change.getChange(), DocumentsChanges_ChangeType_FIELD_CHANGED)

				oldVal := change.getFieldOldValue()
				oldValStr := fmt.Sprintf("%v", oldVal)
				assert.Equal(t, oldValStr, "8")

				newVal := change.getFieldNewValue()
				newValStr := fmt.Sprintf("%v", newVal)
				assert.Equal(t, newValStr, "4")
			}

			{
				change := family1Changes[2]
				assert.Equal(t, change.getFieldName(), "Name")

				assert.Equal(t, change.getChange(), DocumentsChanges_ChangeType_FIELD_CHANGED)

				oldVal := change.getFieldOldValue()
				oldValStr := fmt.Sprintf("%v", oldVal)
				assert.Equal(t, oldValStr, "RavenDB")

				newVal := change.getFieldNewValue()
				newValStr := fmt.Sprintf("%v", newVal)
				assert.Equal(t, newValStr, "Hibernating Rhinos")
			}

			{
				change := family1Changes[3]
				assert.Equal(t, change.getFieldName(), "Age")

				assert.Equal(t, change.getChange(), DocumentsChanges_ChangeType_FIELD_CHANGED)

				oldVal := change.getFieldOldValue()
				oldValStr := fmt.Sprintf("%v", oldVal)
				assert.Equal(t, oldValStr, "4")

				newVal := change.getFieldNewValue()
				newValStr := fmt.Sprintf("%v", newVal)
				assert.Equal(t, newValStr, "8")
			}
		*/
	}

	// TODO: port me
	/*
	   member1 = new Member();
	   member1.setName("Toli");
	   member1.setAge(5);

	   member2 = new Member();
	   member2.setName("Boki");
	   member2.setAge(15);

	   newFamily.setMembers(new Member[]{ member1, member2 });

	   changes = newSession.advanced().whatChanged();
	   assertThat(changes)
	           .hasSize(1);

	   assertThat(changes.get("family/1"))
	           .hasSize(4);

	   assertThat(changes.get("family/1").get(0).getFieldName())
	           .isEqualTo("name");
	   assertThat(changes.get("family/1").get(0).getChange())
	           .isEqualTo(DocumentsChanges.ChangeType.FIELD_CHANGED);
	   assertThat(changes.get("family/1").get(0).getFieldOldValue().toString())
	           .isEqualTo("\"Hibernating Rhinos\"");
	   assertThat(changes.get("family/1").get(0).getFieldNewValue().toString())
	           .isEqualTo("\"Toli\"");

	   assertThat(changes.get("family/1").get(1).getFieldName())
	           .isEqualTo("age");
	   assertThat(changes.get("family/1").get(1).getChange())
	           .isEqualTo(DocumentsChanges.ChangeType.FIELD_CHANGED);
	   assertThat(changes.get("family/1").get(1).getFieldOldValue().toString())
	           .isEqualTo("8");
	   assertThat(changes.get("family/1").get(1).getFieldNewValue().toString())
	           .isEqualTo("5");

	   assertThat(changes.get("family/1").get(2).getFieldName())
	           .isEqualTo("name");
	   assertThat(changes.get("family/1").get(2).getChange())
	           .isEqualTo(DocumentsChanges.ChangeType.FIELD_CHANGED);
	   assertThat(changes.get("family/1").get(2).getFieldOldValue().toString())
	           .isEqualTo("\"RavenDB\"");
	   assertThat(changes.get("family/1").get(2).getFieldNewValue().toString())
	           .isEqualTo("\"Boki\"");

	   assertThat(changes.get("family/1").get(3).getFieldName())
	           .isEqualTo("age");
	   assertThat(changes.get("family/1").get(3).getChange())
	           .isEqualTo(DocumentsChanges.ChangeType.FIELD_CHANGED);
	   assertThat(changes.get("family/1").get(3).getFieldOldValue().toString())
	           .isEqualTo("4");
	   assertThat(changes.get("family/1").get(3).getFieldNewValue().toString())
	           .isEqualTo("15");
	*/
}
func crudTest_crudOperationsWithArrayOfArrays(t *testing.T) {
}
func crudTest_crudCanUpdatePropertyToNull(t *testing.T) {
}
func crudTest_crudCanUpdatePropertyFromNullToObject(t *testing.T) {
}
func TestCrud(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	if useProxy() {
		proxy.ChangeLogFile("trace_crud_go.txt")
	}

	createTestDriver()
	defer deleteTestDriver()

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
