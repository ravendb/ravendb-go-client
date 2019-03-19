package tests

import (
	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
	"testing"
)

func ravendb12790_lazyQueryAgainstMissingIndex(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		document := &Document2{
			Name: "name",
		}
		err = session.Store(document)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	// intentionally not creating the index that we query against
	{
		session := openSessionMust(t, store)
		query := session.QueryIndex("DocumentIndex")
		var results []*Document2
		err = query.GetResults(&results)
		assert.Error(t, err)
		_, ok := err.(*ravendb.IndexDoesNotExistError)
		assert.True(t, ok)

		session.Close()
	}

	{
		session := openSessionMust(t, store)
		lazyQuery, err := session.QueryIndex("DocumentIndex").Lazily()
		assert.NoError(t, err)

		var results []*Document2
		err = lazyQuery.GetValue(&results)
		assert.Error(t, err)
		_, ok := err.(*ravendb.IndexDoesNotExistError)
		assert.True(t, ok)

		session.Close()
	}

}

// note: renamed because conflicts with another Document
type Document2 struct {
	Name string
}

func TestRavenDb12790(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	ravendb12790_lazyQueryAgainstMissingIndex(t, driver)
}
