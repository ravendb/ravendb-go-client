package tests

import (
	"testing"
	"time"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

type CustomType struct {
	ID       string
	Owner    string    `json:"owner"`
	Value    int       `json:"value"`
	Comments []string  `json:"comments"`
	Date     time.Time `json:"date"`
}

func advancedPatchingTestWithVariables(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
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

	patchRequest := &ravendb.PatchRequest{}
	patchRequest.Script = "this.owner = args.v1"
	m := map[string]interface{}{
		"v1": "not-me",
	}
	patchRequest.Values = m
	patchOperation, err := ravendb.NewPatchOperation("customTypes/1", nil, patchRequest, nil, false)
	assert.NoError(t, err)
	err = store.Operations().Send(patchOperation, nil)
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

func advancedPatchingCanCreateDocumentsIfPatchingAppliedByIndex(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
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

		q := session.Advanced().QueryIndex("TestIndex").WaitForNonStaleResults(0)
		var notUsed []*CustomType
		err = q.GetResults(&notUsed)
		assert.NoError(t, err)

		session.Close()
	}

	op2 := ravendb.NewPatchByQueryOperation("FROM INDEX 'TestIndex' WHERE value = 1 update { put('NewItem/3', {'copiedValue': this.value });}")
	operation, err := store.Operations().SendAsync(op2, nil)
	assert.NoError(t, err)

	err = operation.WaitForCompletion()
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		var jsonDoc *map[string]interface{}
		err = session.Load(&jsonDoc, "NewItem/3")
		assert.NoError(t, err)
		m := *jsonDoc
		assert.Equal(t, m["copiedValue"], float64(1))

		session.Close()
	}
}

const SAMPLE_SCRIPT = `this.comments.splice(2, 1);
    this.owner = 'Something new';
    this.value++;
    this.newValue = "err!!";
    this.comments = this.comments.map(function(comment) {
        return (comment == "one") ? comment + " test" : comment;
    });`

func advancedPatchingCanApplyBasicScriptAsPatch(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		test := &CustomType{
			ID:       "someId",
			Owner:    "bob",
			Value:    12143,
			Comments: []string{"one", "two", "seven"},
		}

		err = session.Store(test)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}
	req := &ravendb.PatchRequest{
		Script: SAMPLE_SCRIPT,
	}

	op, err := ravendb.NewPatchOperation("someId", nil, req, nil, false)
	assert.NoError(t, err)
	err = store.Operations().Send(op, nil)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		var result *CustomType
		err = session.Load(&result, "someId")

		assert.Equal(t, result.Owner, "Something new")
		assert.Equal(t, len(result.Comments), 2)
		assert.Equal(t, result.Comments[0], "one test")
		assert.Equal(t, result.Comments[1], "two")
		assert.Equal(t, result.Value, 12144)

		session.Close()
	}
}

func advancedPatchingCanDeserializeModifiedDocument(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	customType := &CustomType{
		Owner: "somebody@somewhere.com",
	}
	{
		session := openSessionMust(t, store)

		err = session.StoreWithID(customType, "doc")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	req := &ravendb.PatchRequest{
		Script: "this.owner = '123';",
	}
	patch1, err := ravendb.NewPatchOperation("doc", nil, req, nil, false)
	assert.NoError(t, err)
	patchResult, err := store.Operations().SendPatchOperation(patch1, nil)
	assert.NoError(t, err)
	var result *CustomType
	err = patchResult.GetResult(&result)
	assert.NoError(t, err)

	assert.Equal(t, patchResult.Status, ravendb.PatchStatusPatched)
	assert.Equal(t, result.Owner, "123")

	patch2, err := ravendb.NewPatchOperation("doc", nil, req, nil, false)
	assert.NoError(t, err)
	patchResult, err = store.Operations().SendPatchOperation(patch2, nil)
	assert.NoError(t, err)
	result = nil
	err = patchResult.GetResult(&result)
	assert.NoError(t, err)

	assert.Equal(t, patchResult.Status, ravendb.PatchStatusNotModified)
	assert.Equal(t, result.Owner, "123")
}

func TestAdvancedPatching(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	advancedPatchingTestWithVariables(t, driver)
	advancedPatchingCanCreateDocumentsIfPatchingAppliedByIndex(t, driver)

	// TODO: order doesn't match Java
	advancedPatchingCanApplyBasicScriptAsPatch(t, driver)
	advancedPatchingCanDeserializeModifiedDocument(t, driver)
}
