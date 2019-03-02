package tests

import (
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func suggestionsLazyUsingLinq(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	indexDefinition := ravendb.NewIndexDefinition()
	indexDefinition.Name = "test"

	indexDefinition.Maps = []string{"from doc in docs.User4s select new { doc.name }"}
	indexFieldOptions := ravendb.NewIndexFieldOptions()
	indexFieldOptions.Suggestions = true
	indexDefinition.Fields["name"] = indexFieldOptions

	err = store.Maintenance().Send(ravendb.NewPutIndexesOperation(indexDefinition))
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)

		user1 := &User4{
			Name: "Ayende",
		}
		err = session.Store(user1)
		assert.NoError(t, err)

		user2 := &User4{
			Name: "Oren",
		}
		err = session.Store(user2)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	err = driver.waitForIndexing(store, "", 0)
	assert.NoError(t, err)

	{
		s := openSessionMust(t, store)
		oldRequests := s.Advanced().GetNumberOfRequests()

		q := s.QueryIndex("test")
		x := ravendb.NewSuggestionBuilder()
		x = x.ByField("name", "Owen")
		q2 := q.SuggestUsing(x.GetSuggestion())

		suggestionQueryResult, err := q2.ExecuteLazy()
		assert.NoError(t, err)
		assert.Equal(t, oldRequests, s.Advanced().GetNumberOfRequests())

		result := map[string]*ravendb.SuggestionResult{}
		err = suggestionQueryResult.GetValue(&result)
		assert.NoError(t, err)
		suggestions := result["name"].Suggestions
		assert.Equal(t, len(suggestions), 1)
		assert.Equal(t, suggestions[0], "oren")

		assert.Equal(t, oldRequests+1, s.Advanced().GetNumberOfRequests())

		s.Close()
	}
}

func TestSuggestionsLazy(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	suggestionsLazyUsingLinq(t, driver)
}
