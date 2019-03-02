package tests

import (
	"testing"
	"time"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func simpleMultiMap_canCreateMultiMapIndex(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewCatsAndDogs()
	err = index.Execute(store, nil, "")
	assert.NoError(t, err)

	op := ravendb.NewGetIndexOperation("CatsAndDogs")
	err = store.Maintenance().Send(op)
	indexDefinition := op.Command.Result
	assert.Equal(t, len(indexDefinition.Maps), 2)
}

func simpleMultiMap_canQueryUsingMultiMap(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewCatsAndDogs()
	err = index.Execute(store, nil, "")
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		cat := &Cat{Name: "Tom"}
		dog := &Dog{Name: "Oscar"}

		err = session.Store(cat)
		assert.NoError(t, err)
		err = session.Store(dog)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		// Note: Go doesn't support interfaces like Java. We can only
		// query a single type, not an interface
		var haveNames []*Dog
		q := session.QueryIndex(index.IndexName)
		q = q.WaitForNonStaleResults(time.Second * 10)
		q = q.OrderBy("name")
		err = q.GetResults(&haveNames)
		assert.NoError(t, err)

		assert.Equal(t, len(haveNames), 2)

		assert.Equal(t, haveNames[0].Name, "Oscar")
		assert.Equal(t, haveNames[1].Name, "Tom")

		session.Close()
	}
}

// Note: in Go IndexCreationTask covers functionality of
// AbstractMultiMapIndexCreationTask
func NewCatsAndDogs() *ravendb.IndexCreationTask {
	res := ravendb.NewIndexCreationTask("CatsAndDogs")
	res.Maps = []string{
		"from cat in docs.Cats select new { cat.name }",
		"from dog in docs.Dogs select new { dog.name }",
	}
	return res
}

type Cat struct {
	Name string `json:"name"`
}

// Note: re-using a Dog structure from other test, which also has Name field

func TestSimpleMultiMap(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	simpleMultiMap_canCreateMultiMapIndex(t, driver)
	simpleMultiMap_canQueryUsingMultiMap(t, driver)
}
