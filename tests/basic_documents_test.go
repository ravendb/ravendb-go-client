package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

func basicDocuments_canChangeDocumentCollectionWithDeleteAndSave(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	documentId := "users/1"
	{
		session := openSessionMust(t, store)
		user := NewUser()
		user.setName("Grisha")

		err = session.StoreWithID(user, documentId)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		session := openSessionMust(t, store)
		err = session.Delete(documentId)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		session := openSessionMust(t, store)
		userI, err := session.Load(ravendb.GetTypeOf(&User{}), documentId)
		assert.NoError(t, err)
		assert.Nil(t, userI)
		session.Close()
	}

	{
		session := openSessionMust(t, store)
		person := &Person{}
		person.setName("Grisha")
		err = session.StoreWithID(person, documentId)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}
}

func basicDocuments_get(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	dummy := ravendb.ValueToTree(NewUser())
	delete(dummy, "ID")

	{
		session := openSessionMust(t, store)
		user1 := NewUser()
		user1.setName("Fitzchak")

		user2 := NewUser()
		user2.setName("Arek")

		err = session.StoreWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.StoreWithID(user2, "users/2")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}
	requestExecutor := store.GetRequestExecutor()
	getDocumentsCommand := ravendb.NewGetDocumentsCommand([]string{"users/1", "users/2"}, nil, false)
	err = requestExecutor.ExecuteCommand(getDocumentsCommand)
	assert.NoError(t, err)
	docs := getDocumentsCommand.Result
	assert.Equal(t, len(docs.GetResults()), 2)
	doc1 := docs.GetResults()[0]
	doc2 := docs.GetResults()[1]

	assert.NotNil(t, doc1)
	doc1Properties := ravendb.FieldNames(doc1)
	assert.True(t, ravendb.StringArrayContains(doc1Properties, "@metadata"))
	assert.Equal(t, len(doc1Properties), len(dummy)+1) // +1 for @metadata

	assert.NotNil(t, doc2)
	doc2Properties := ravendb.FieldNames(doc2)
	assert.True(t, ravendb.StringArrayContains(doc2Properties, "@metadata"))
	assert.Equal(t, len(doc2Properties), len(dummy)+1) // +1 for @metadata

	{
		session := openSessionMust(t, store)
		etojs := session.GetEntityToJson()
		user1I := etojs.ConvertToEntity(ravendb.GetTypeOf(&User{}), "users/1", doc1)
		user1 := user1I.(*User)

		user2I := etojs.ConvertToEntity(ravendb.GetTypeOf(&User{}), "users/2", doc2)
		user2 := user2I.(*User)
		assert.Equal(t, *user1.GetName(), "Fitzchak")
		assert.Equal(t, *user2.GetName(), "Arek")
		session.Close()
	}
	getDocumentsCommand = ravendb.NewGetDocumentsCommand([]string{"users/1", "users/2"}, nil, true)
	err = requestExecutor.ExecuteCommand(getDocumentsCommand)
	docs = getDocumentsCommand.Result
	assert.Equal(t, len(docs.GetResults()), 2)
	doc1 = docs.GetResults()[0]
	doc2 = docs.GetResults()[1]

	assert.NotNil(t, doc1)
	doc1Properties = ravendb.FieldNames(doc1)
	assert.True(t, ravendb.StringArrayContains(doc1Properties, "@metadata"))
	assert.Equal(t, len(doc1Properties), 1)

	assert.NotNil(t, doc1)
	doc2Properties = ravendb.FieldNames(doc2)
	assert.True(t, ravendb.StringArrayContains(doc2Properties, "@metadata"))
	assert.Equal(t, len(doc2Properties), 1)
}

func TestBasicDocuments(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	basicDocuments_get(t)
	basicDocuments_canChangeDocumentCollectionWithDeleteAndSave(t)
}
