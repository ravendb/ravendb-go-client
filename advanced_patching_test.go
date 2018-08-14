package ravendb

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type CustomType struct {
	ID       string
	Owner    string    `json:"owner"`
	Value    int       `json:"value"`
	Comments []string  `json:"comments"`
	Date     time.Time `json:"date"`
}

func advancedPatching_testWithVariables(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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

	patchRequest := NewPatchRequest()
	patchRequest.SetScript("this.owner = args.v1")
	m := map[string]Object{
		"v1": "not-me",
	}
	patchRequest.SetValues(m)
	patchOperation := NewPatchOperation("customTypes/1", nil, patchRequest, nil, false)
	err = store.Operations().send(patchOperation)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		loadedI, err := session.Load(GetTypeOf(&CustomType{}), "customTypes/1")
		assert.NoError(t, err)
		loaded := loadedI.(*CustomType)
		assert.Equal(t, loaded.Owner, "not-me")

		session.Close()
	}
}

func advancedPatching_canCreateDocumentsIfPatchingAppliedByIndex(t *testing.T) {

	var err error
	store := getDocumentStoreMust(t)
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

	def1 := NewIndexDefinition()
	def1.setName("TestIndex")
	def1.setMaps(NewStringSetFromStrings("from doc in docs.CustomTypes select new { doc.value }"))

	op := NewPutIndexesOperation(def1)
	err = store.Maintenance().send(op)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		q := session.Advanced().DocumentQueryAll(GetTypeOf(&CustomType{}), "TestIndex", "", false)
		q = q.waitForNonStaleResults(0)
		_, err = q.toList()
		assert.NoError(t, err)

		session.Close()
	}

	op2 := NewPatchByQueryOperation("FROM INDEX 'TestIndex' WHERE value = 1 update { put('NewItem/3', {'copiedValue': this.value });}")
	operation, err := store.Operations().sendAsync(op2)
	assert.NoError(t, err)

	operation.waitForCompletion()

	{
		session := openSessionMust(t, store)

		jsonDocument, err := session.Load(GetTypeOf(ObjectNode{}), "NewItem/3")
		assert.NoError(t, err)
		jsonDoc := jsonDocument.(ObjectNode)
		assert.Equal(t, jsonDoc["copiedValue"], "1")

		session.Close()
	}
}

func TestAdvancedPatching(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	advancedPatching_testWithVariables(t)
	if gEnableFailingTests {
		// TODO: fails because documentsByEntity cannot handle map[string]interface{}
		advancedPatching_canCreateDocumentsIfPatchingAppliedByIndex(t)
	}
}
