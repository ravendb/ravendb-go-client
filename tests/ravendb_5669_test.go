package tests

import (
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func ravendb5669WorkingTestWithDifferentSearchTermOrder(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewAnimalIndex()
	err = store.ExecuteIndex(index, "")
	assert.NoError(t, err)

	ravendb5669storeAnimals(t, store, driver)

	{
		session := openSessionMust(t, store)

		var results []*Animal
		query, err := session.Advanced().QueryIndex(index.IndexName)
		assert.NoError(t, err)

		query.OpenSubclause()

		query = query.WhereEquals("type", "Cat")
		query = query.OrElse()
		query = query.Search("name", "Peter*")
		query = query.AndAlso()
		query = query.Search("name", "Pan*")

		query.CloseSubclause()

		err = query.GetResults(&results)
		assert.NoError(t, err)
		assert.Equal(t, len(results), 1)

		session.Close()
	}
}

func ravendb5669workingTestWithSubclause(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewAnimalIndex()
	err = store.ExecuteIndex(index, "")
	assert.NoError(t, err)

	ravendb5669storeAnimals(t, store, driver)

	{
		session := openSessionMust(t, store)

		query, err := session.Advanced().QueryIndex(index.IndexName)
		assert.NoError(t, err)

		query.OpenSubclause()

		query = query.WhereEquals("type", "Cat")
		query = query.OrElse()

		query.OpenSubclause()

		query = query.Search("name", "Pan*")
		query = query.AndAlso()
		query = query.Search("name", "Peter*")
		query = query.CloseSubclause()

		query.CloseSubclause()

		var results []*Animal
		err = query.GetResults(&results)
		assert.NoError(t, err)
		assert.Equal(t, len(results), 1)

		session.Close()
	}
}

func ravendb5669storeAnimals(t *testing.T, store *ravendb.DocumentStore, driver *RavenTestDriver) {
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

	driver.waitForIndexing(store, store.GetDatabase(), 0)
}

type Animal struct {
	Type string `json:"type"`
	Name string `json:"name"`
}

func NewAnimalIndex() *ravendb.IndexCreationTask {
	res := ravendb.NewIndexCreationTask("Animal_Index")
	res.Map = "from animal in docs.Animals select new { name = animal.name, type = animal.type }"

	res.Analyze("name", "StandardAnalyzer")
	res.Index("name", ravendb.FieldIndexingSearch)
	return res
}

func TestRavenDB5669(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches the order of Java tests
	ravendb5669workingTestWithSubclause(t, driver)
	ravendb5669WorkingTestWithDifferentSearchTermOrder(t, driver)
}
