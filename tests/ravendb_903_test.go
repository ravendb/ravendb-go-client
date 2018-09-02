package tests

import (
	"reflect"
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

// unique name that doesn't conflict with Product in hi_lo_test.go
type Product2 struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func ravendb_903_test1(t *testing.T) {

	fn := func(session *ravendb.DocumentSession, index *ravendb.AbstractIndexCreationTask) *ravendb.IDocumentQuery {
		q := session.Advanced().DocumentQueryInIndexOld(reflect.TypeOf(&Product2{}), index)
		q = q.Search("description", "Hello")
		q = q.Intersect()
		q = q.WhereEquals("name", "Bar")
		return q
	}
	ravendb_903_doTest(t, fn)
}

func ravendb_903_test2(t *testing.T) {
	fn := func(session *ravendb.DocumentSession, index *ravendb.AbstractIndexCreationTask) *ravendb.IDocumentQuery {
		q := session.Advanced().DocumentQueryInIndexOld(reflect.TypeOf(&Product2{}), index)
		q = q.WhereEquals("name", "Bar")
		q = q.Intersect()
		q = q.Search("description", "Hello")
		return q
	}
	ravendb_903_doTest(t, fn)

}

func ravendb_903_doTest(t *testing.T, queryFunction func(*ravendb.DocumentSession, *ravendb.AbstractIndexCreationTask) *ravendb.IDocumentQuery) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewTestIndex()
	err = store.ExecuteIndex(index)
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

	gRavenTestDriver.waitForIndexing(store, "", 0)

	{
		var products []*Product2
		session := openSessionMust(t, store)
		query := queryFunction(session, index)
		err = query.ToList(&products)
		assert.NoError(t, err)
		assert.Equal(t, len(products), 1)

		session.Close()
	}
}

func NewTestIndex() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("TestIndex")
	res.Map = "from product in docs.Product2s select new { product.name, product.description }"
	res.Index("description", ravendb.FieldIndexing_SEARCH)
	return res
}

func TestRavenDB903(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches the order of Java tests
	ravendb_903_test1(t)
	ravendb_903_test2(t)
}
