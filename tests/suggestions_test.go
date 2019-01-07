package tests

import (
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

// avoids conflict with other User classes
type User4 struct {
	ID    string
	Name  string `json:"name"`
	Email string `json:"email"`
}

func suggestionsSetup(t *testing.T, driver *RavenTestDriver, store *ravendb.DocumentStore) {
	var err error
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

		user2 := &User4{
			Name: "Oren",
		}

		user3 := &User4{
			Name: "John Steinbeck",
		}

		err = session.Store(user1)
		assert.NoError(t, err)
		err = session.Store(user2)
		assert.NoError(t, err)
		err = session.Store(user3)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	driver.waitForIndexing(store, "", 0)
}

func suggestionsExactMatch(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	suggestionsSetup(t, driver, store)
	{
		session := openSessionMust(t, store)

		options := ravendb.NewSuggestionOptions()
		options.PageSize = 10

		queryIndex := &ravendb.Query{
			IndexName: "test",
		}
		q := session.QueryWithQuery(queryIndex)
		fn := func(x ravendb.ISuggestionBuilder) {
			x.ByField("name", "Oren").WithOptions(options)
		}
		q2 := q.SuggestUsingBuilder(fn)
		suggestionQueryResult, err := q2.Execute()
		assert.NoError(t, err)

		assert.Equal(t, len(suggestionQueryResult["name"].Suggestions), 0)

		session.Close()
	}
}

func suggestionsUsingLinq(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	suggestionsSetup(t, driver, store)
	{
		s := openSessionMust(t, store)

		queryIndex := &ravendb.Query{
			IndexName: "test",
		}
		q := s.QueryWithQuery(queryIndex)
		fn := func(x ravendb.ISuggestionBuilder) {
			x.ByField("name", "Owen")
		}
		q2 := q.SuggestUsingBuilder(fn)
		suggestionQueryResult, err := q2.Execute()
		assert.NoError(t, err)

		su := suggestionQueryResult["name"].Suggestions
		assert.Equal(t, len(su), 1)
		assert.Equal(t, su[0], "oren")

		s.Close()
	}
}

func suggestionsUsingLinqWithOptions(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	suggestionsSetup(t, driver, store)
	{
		s := openSessionMust(t, store)

		options := ravendb.NewSuggestionOptions()
		options.Accuracy = 0.4
		queryIndex := &ravendb.Query{
			IndexName: "test",
		}
		q := s.QueryWithQuery(queryIndex)
		fn := func(x ravendb.ISuggestionBuilder) {
			x.ByField("name", "Owen").WithOptions(options)
		}
		q2 := q.SuggestUsingBuilder(fn)
		suggestionQueryResult, err := q2.Execute()
		assert.NoError(t, err)

		su := suggestionQueryResult["name"].Suggestions
		assert.Equal(t, len(su), 1)
		assert.Equal(t, su[0], "oren")

		s.Close()
	}
}

func suggestionsUsingLinqMultipleWords(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	suggestionsSetup(t, driver, store)
	{
		s := openSessionMust(t, store)

		options := ravendb.NewSuggestionOptions()
		options.Accuracy = 0.4
		options.Distance = ravendb.StringDistanceLevenshtein

		queryIndex := &ravendb.Query{
			IndexName: "test",
		}
		q := s.QueryWithQuery(queryIndex)
		fn := func(x ravendb.ISuggestionBuilder) {
			x.ByField("name", "John Steinback").WithOptions(options)
		}
		q2 := q.SuggestUsingBuilder(fn)
		suggestionQueryResult, err := q2.Execute()
		assert.NoError(t, err)

		su := suggestionQueryResult["name"].Suggestions
		assert.Equal(t, len(su), 1)
		assert.Equal(t, su[0], "john steinbeck")

		s.Close()
	}
}

func suggestionsWithTypo(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	suggestionsSetup(t, driver, store)
	{
		s := openSessionMust(t, store)

		options := ravendb.NewSuggestionOptions()
		options.Accuracy = 0.2
		options.PageSize = 10
		options.Distance = ravendb.StringDistanceLevenshtein

		queryIndex := &ravendb.Query{
			IndexName: "test",
		}
		q := s.QueryWithQuery(queryIndex)
		fn := func(x ravendb.ISuggestionBuilder) {
			x.ByField("name", "Oern").WithOptions(options)
		}
		q2 := q.SuggestUsingBuilder(fn)
		suggestionQueryResult, err := q2.Execute()
		assert.NoError(t, err)

		su := suggestionQueryResult["name"].Suggestions
		assert.Equal(t, len(su), 1)
		assert.Equal(t, su[0], "oren")

		s.Close()
	}
}

func NewUsers4ByName() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("NewUsers_ByName")
	res.Map = "from u in docs.User4s select new { u.name }"

	res.Index("name", ravendb.FieldIndexingSearch)

	res.IndexSuggestions = append(res.IndexSuggestions, "name")

	res.Store("name", ravendb.FieldStorageYes)

	return res
}

func suggestionsCanGetSuggestions(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsers4ByName()
	err = index.Execute(store)
	assert.NoError(t, err)

	{
		s := openSessionMust(t, store)

		user1 := &User4{
			Name: "John Smith",
		}
		err = s.StoreWithID(user1, "users/1")
		assert.NoError(t, err)

		user2 := &User4{
			Name: "Jack Johnson",
		}
		err = s.StoreWithID(user2, "users/2")
		assert.NoError(t, err)

		user3 := &User4{
			Name: "Robery Jones",
		}
		err = s.StoreWithID(user3, "users/3")
		assert.NoError(t, err)

		user4 := &User4{
			Name: "David Jones",
		}
		err = s.StoreWithID(user4, "users/4")
		assert.NoError(t, err)

		err = s.SaveChanges()
		assert.NoError(t, err)

		s.Close()
	}

	driver.waitForIndexing(store, "", 0)

	{
		session := openSessionMust(t, store)

		options := ravendb.NewSuggestionOptions()
		options.Accuracy = 0.4
		options.PageSize = 5
		options.Distance = ravendb.StringDistanceJaroWinkler
		options.SortMode = ravendb.SuggestionSortModePopularity

		q := session.QueryInIndex(index)
		fn := func(x ravendb.ISuggestionBuilder) {
			x.ByField("name", "johne", "davi").WithOptions(options)
		}
		q2 := q.SuggestUsingBuilder(fn)
		suggestionQueryResult, err := q2.Execute()
		assert.NoError(t, err)

		su := suggestionQueryResult["name"].Suggestions
		assert.Equal(t, len(su), 5)
		ok := stringArrayContainsSequence(su, []string{"john", "jones", "johnson", "david", "jack"})
		assert.True(t, ok)

		session.Close()
	}
}

func TestSuggestions(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches the order of Java tests
	suggestionsCanGetSuggestions(t, driver)
	suggestionsUsingLinqMultipleWords(t, driver)
	suggestionsWithTypo(t, driver)
	suggestionsUsingLinq(t, driver)
	suggestionsUsingLinqWithOptions(t, driver)
	suggestionsExactMatch(t, driver)
}
