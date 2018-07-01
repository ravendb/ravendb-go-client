package ravendb

import (
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

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
