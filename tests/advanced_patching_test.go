package tests

import (
	"reflect"
	"testing"
	"time"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

type CustomType struct {
	ID       string
	Owner    string    `json:"owner"`
	Value    int       `json:"value"`
	Comments []string  `json:"comments"`
	Date     time.Time `json:"date"`
}

func advancedPatching_testWithVariables(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		customType := &CustomType{
			Owner: "me",
		}
		err = session.StoreWithID(customType, "customTypes/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	patchRequest := ravendb.NewPatchRequest()
	patchRequest.SetScript("this.owner = args.v1")
	m := map[string]interface{}{
		"v1": "not-me",
	}
	patchRequest.SetValues(m)
	patchOperation := ravendb.NewPatchOperation("customTypes/1", nil, patchRequest, nil, false)
	err = store.Operations().Send(patchOperation)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		var loaded *CustomType
		err = session.Load(&loaded, "customTypes/1")
		assert.NoError(t, err)
		assert.Equal(t, loaded.Owner, "not-me")

		session.Close()
	}
}

func advancedPatching_canCreateDocumentsIfPatchingAppliedByIndex(t *testing.T, driver *RavenTestDriver) {

	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		newSession := openSessionMust(t, store)
		type1 := &CustomType{
			ID:    "Item/1",
			Value: 1,
		}

		type2 := &CustomType{
			ID:    "Item/2",
			Value: 2,
		}

		err = newSession.Store(type1)
		assert.NoError(t, err)
		err = newSession.Store(type2)
		assert.NoError(t, err)
		err = newSession.SaveChanges()
		assert.NoError(t, err)

		newSession.Close()
	}

	def1 := ravendb.NewIndexDefinition()
	def1.Name = "TestIndex"
	def1.Maps = []string{"from doc in docs.CustomTypes select new { doc.value }"}

	op := ravendb.NewPutIndexesOperation(def1)
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		var notUsed []*CustomType
		q := session.Advanced().DocumentQueryAllOld(reflect.TypeOf(&CustomType{}), "TestIndex", "", false)
		q = q.WaitForNonStaleResults(0)
		err = q.ToList(&notUsed)
		assert.NoError(t, err)

		session.Close()
	}

	op2 := ravendb.NewPatchByQueryOperation("FROM INDEX 'TestIndex' WHERE value = 1 update { put('NewItem/3', {'copiedValue': this.value });}")
	operation, err := store.Operations().SendAsync(op2)
	assert.NoError(t, err)

	operation.WaitForCompletion()

	{
		session := openSessionMust(t, store)

		var jsonDoc ravendb.ObjectNode
		err = session.Load(&jsonDoc, "NewItem/3")
		assert.NoError(t, err)
		assert.Equal(t, jsonDoc["copiedValue"], float64(1))

		session.Close()
	}
}

func TestAdvancedPatching(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	advancedPatching_testWithVariables(t, driver)
	advancedPatching_canCreateDocumentsIfPatchingAppliedByIndex(t, driver)
}
