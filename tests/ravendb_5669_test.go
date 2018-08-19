package tests

import (
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func ravendb5669_workingTestWithDifferentSearchTermOrder(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewAnimal_Index()
	err = store.ExecuteIndex(index)
	assert.NoError(t, err)

	ravendb5669_storeAnimals(t, store)

	{
		session := openSessionMust(t, store)

		query := session.Advanced().DocumentQueryInIndex(ravendb.GetTypeOf(&Animal{}), index)

		query.OpenSubclause()

		query = query.WhereEquals("type", "Cat")
		query = query.OrElse()
		query = query.Search("name", "Peter*")
		query = query.AndAlso()
		query = query.Search("name", "Pan*")

		query.CloseSubclause()

		results, err := query.ToList()
		assert.NoError(t, err)
		assert.Equal(t, len(results), 1)

		session.Close()
	}
}

func ravendb5669_workingTestWithSubclause(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewAnimal_Index()
	err = store.ExecuteIndex(index)
	assert.NoError(t, err)

	ravendb5669_storeAnimals(t, store)

	{
		session := openSessionMust(t, store)

		query := session.Advanced().DocumentQueryInIndex(ravendb.GetTypeOf(&Animal{}), index)

		query.OpenSubclause()

		query = query.WhereEquals("type", "Cat")
		query = query.OrElse()

		query.OpenSubclause()

		query = query.Search("name", "Pan*")
		query = query.AndAlso()
		query = query.Search("name", "Peter*")
		query = query.CloseSubclause()

		query.CloseSubclause()

		results, err := query.ToList()
		assert.NoError(t, err)
		assert.Equal(t, len(results), 1)

		session.Close()
	}
}

func ravendb5669_storeAnimals(t *testing.T, store *ravendb.DocumentStore) {
	var err error

	{
		session := openSessionMust(t, store)

		animal1 := &Animal{
			Name: "Peter Pan",
			Type: "Dog",
		}

		animal2 := &Animal{
			Name: "Peter Poo",
			Type: "Dog",
		}

		animal3 := &Animal{
			Name: "Peter Foo",
			Type: "Dog",
		}

		err = session.Store(animal1)
		assert.NoError(t, err)
		err = session.Store(animal2)
		assert.NoError(t, err)
		err = session.Store(animal3)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
}

type Animal struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

func NewAnimal_Index() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("Animal_Index")
	res.Map = "from animal in docs.Animals select new { name = animal.name, type = animal.type }"

	res.Analyze("name", "StandardAnalyzer")
	res.Index("name", ravendb.FieldIndexing_SEARCH)
	return res
}

func TestRavenDB5669(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches the order of Java tests
	ravendb5669_workingTestWithSubclause(t)
	ravendb5669_workingTestWithDifferentSearchTermOrder(t)
}
