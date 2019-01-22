package tests

import (
	"reflect"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func basicDocumentscanChangeDocumentCollectionWithDeleteAndSave(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	documentID := "users/1"
	{
		session := openSessionMust(t, store)
		user := &User{}
		user.setName("Grisha")

		err = session.StoreWithID(user, documentID)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		session := openSessionMust(t, store)
		err = session.Delete(documentID)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		session := openSessionMust(t, store)
		var user *User
		err = session.Load(&user, documentID)
		assert.NoError(t, err)
		assert.Nil(t, user)
		session.Close()
	}

	{
		session := openSessionMust(t, store)
		person := &Person{}
		person.Name = "Grisha"
		err = session.StoreWithID(person, documentID)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}
}

func basicDocumentsGet(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	dummy := ravendb.ValueToTree(&User{})
	delete(dummy, "ID")

	{
		session := openSessionMust(t, store)
		user1 := &User{}
		user1.setName("Fitzchak")

		user2 := &User{}
		user2.setName("Arek")

		err = session.StoreWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.StoreWithID(user2, "users/2")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}
	requestExecutor := store.GetRequestExecutor("")
	getDocumentsCommand := ravendb.NewGetDocumentsCommand([]string{"users/1", "users/2"}, nil, false)
	err = requestExecutor.ExecuteCommand(getDocumentsCommand)
	assert.NoError(t, err)
	docs := getDocumentsCommand.Result
	assert.Equal(t, len(docs.Results), 2)
	doc1 := docs.Results[0]
	doc2 := docs.Results[1]

	assert.NotNil(t, doc1)
	doc1Properties := ravendb.FieldNames(doc1)
	assert.True(t, stringArrayContains(doc1Properties, "@metadata"))
	assert.Equal(t, len(doc1Properties), len(dummy)+1) // +1 for @metadata

	assert.NotNil(t, doc2)
	doc2Properties := ravendb.FieldNames(doc2)
	assert.True(t, stringArrayContains(doc2Properties, "@metadata"))
	assert.Equal(t, len(doc2Properties), len(dummy)+1) // +1 for @metadata

	{
		session := openSessionMust(t, store)
		etojs := session.GetEntityToJSON()
		user1I, err := etojs.ConvertToEntity(reflect.TypeOf(&User{}), "users/1", doc1)
		assert.NoError(t, err)
		user1 := user1I.(*User)

		user2I, err := etojs.ConvertToEntity(reflect.TypeOf(&User{}), "users/2", doc2)
		assert.NoError(t, err)
		user2 := user2I.(*User)
		assert.Equal(t, *user1.Name, "Fitzchak")
		assert.Equal(t, *user2.Name, "Arek")
		session.Close()
	}
	getDocumentsCommand = ravendb.NewGetDocumentsCommand([]string{"users/1", "users/2"}, nil, true)
	err = requestExecutor.ExecuteCommand(getDocumentsCommand)
	assert.NoError(t, err)
	docs = getDocumentsCommand.Result
	assert.Equal(t, len(docs.Results), 2)
	doc1 = docs.Results[0]
	doc2 = docs.Results[1]

	assert.NotNil(t, doc1)
	doc1Properties = ravendb.FieldNames(doc1)
	assert.True(t, stringArrayContains(doc1Properties, "@metadata"))
	assert.Equal(t, len(doc1Properties), 1)

	assert.NotNil(t, doc1)
	doc2Properties = ravendb.FieldNames(doc2)
	assert.True(t, stringArrayContains(doc2Properties, "@metadata"))
	assert.Equal(t, len(doc2Properties), 1)
}

func TestBasicDocuments(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	basicDocumentsGet(t, driver)
	basicDocumentscanChangeDocumentCollectionWithDeleteAndSave(t, driver)
}
