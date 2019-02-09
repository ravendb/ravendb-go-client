package tests

import (
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

// unique name that doesn't conflict with Product in hi_lo_test.go
type Product2 struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func ravendb903Test1(t *testing.T, driver *RavenTestDriver) {

	fn := func(session *ravendb.DocumentSession, index *ravendb.AbstractIndexCreationTask) *ravendb.DocumentQuery {
		q, err := session.Advanced().QueryIndex(index.IndexName)
		assert.NoError(t, err)
		q = q.Search("description", "Hello")
		q = q.Intersect()
		q = q.WhereEquals("name", "Bar")
		return q
	}
	ravendb903DoTest(t, driver, fn)
}

func ravendb903Test2(t *testing.T, driver *RavenTestDriver) {
	fn := func(session *ravendb.DocumentSession, index *ravendb.AbstractIndexCreationTask) *ravendb.DocumentQuery {
		q, err := session.Advanced().QueryIndex(index.IndexName)
		assert.NoError(t, err)
		q = q.WhereEquals("name", "Bar")
		q = q.Intersect()
		q = q.Search("description", "Hello")
		return q
	}
	ravendb903DoTest(t, driver, fn)

}

func ravendb903DoTest(t *testing.T, driver *RavenTestDriver, queryFunction func(*ravendb.DocumentSession, *ravendb.AbstractIndexCreationTask) *ravendb.DocumentQuery) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewTestIndex()
	err = store.ExecuteIndex(index, "")
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		product1 := &Product2{
			Name:        "Foo",
			Description: "Hello World",
		}

		product2 := &Product2{
			Name:        "Bar",
			Description: "Hello World",
		}

		product3 := &Product2{
			Name:        "Bar",
			Description: "Goodbye World",
		}

		err = session.Store(product1)
		assert.NoError(t, err)
		err = session.Store(product2)
		assert.NoError(t, err)
		err = session.Store(product3)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	driver.waitForIndexing(store, "", 0)

	{
		var products []*Product2
		session := openSessionMust(t, store)
		query := queryFunction(session, index)
		err = query.GetResults(&products)
		assert.NoError(t, err)
		assert.Equal(t, len(products), 1)

		session.Close()
	}
}

func NewTestIndex() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("TestIndex")
	res.Map = "from product in docs.Product2s select new { product.name, product.description }"
	res.Index("description", ravendb.FieldIndexingSearch)
	return res
}

func TestRavenDB903(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches the order of Java tests
	ravendb903Test1(t, driver)
	ravendb903Test2(t, driver)
}
