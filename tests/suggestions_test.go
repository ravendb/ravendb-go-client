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

func suggestions_setup(t *testing.T, store *ravendb.IDocumentStore) {
	var err error
	indexDefinition := ravendb.NewIndexDefinition()
	indexDefinition.Name = "test"
	indexDefinition.Maps = []string{"from doc in docs.User4s select new { doc.name }"}
	indexFieldOptions := ravendb.NewIndexFieldOptions()
	indexFieldOptions.Suggestions = true
	indexDefinition.Fields["name"] = indexFieldOptions

	err = store.Maintenance().Send(ravendb.NewPutIndexesOperation(indexDefinition))

	{
		session := openSessionMust(t, store)

		user1 := User4{
			Name: "Ayende",
		}

		user2 := User4{
			Name: "Oren",
		}

		user3 := User4{
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

	gRavenTestDriver.waitForIndexing(store, "", 0)
}

func suggestions_exactMatch(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	suggestions_setup(t, store)
	{
		session := openSessionMust(t, store)

		options := ravendb.NewSuggestionOptions()
		options.PageSize = 10

		q := session.QueryWithQueryOld(ravendb.GetTypeOf(&User4{}), ravendb.Query_index("test"))
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

func suggestions_usingLinq(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	suggestions_setup(t, store)
	{
		s := openSessionMust(t, store)

		q := s.QueryWithQueryOld(ravendb.GetTypeOf(&User4{}), ravendb.Query_index("test"))
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

func suggestions_usingLinq_WithOptions(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	suggestions_setup(t, store)
	{
		s := openSessionMust(t, store)

		options := ravendb.NewSuggestionOptions()
		options.Accuracy = 0.4
		q := s.QueryWithQueryOld(ravendb.GetTypeOf(&User4{}), ravendb.Query_index("test"))
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

func suggestions_usingLinq_Multiple_words(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	suggestions_setup(t, store)
	{
		s := openSessionMust(t, store)

		options := ravendb.NewSuggestionOptions()
		options.Accuracy = 0.4
		options.Distance = ravendb.StringDistanceTypes_LEVENSHTEIN

		q := s.QueryWithQueryOld(ravendb.GetTypeOf(&User4{}), ravendb.Query_index("test"))
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

func suggestions_withTypo(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	suggestions_setup(t, store)
	{
		s := openSessionMust(t, store)

		options := ravendb.NewSuggestionOptions()
		options.Accuracy = 0.2
		options.PageSize = 10
		options.Distance = ravendb.StringDistanceTypes_LEVENSHTEIN

		q := s.QueryWithQueryOld(ravendb.GetTypeOf(&User4{}), ravendb.Query_index("test"))
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

func NewUsers4_ByName() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("NewUsers_ByName")
	res.Map = "from u in docs.User4s select new { u.name }"

	res.Index("name", ravendb.FieldIndexing_SEARCH)

	res.IndexSuggestions = append(res.IndexSuggestions, "name")

	res.Store("name", ravendb.FieldStorage_YES)

	return res
}

func suggestions_canGetSuggestions(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsers4_ByName()
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

	gRavenTestDriver.waitForIndexing(store, "", 0)

	{
		session := openSessionMust(t, store)

		options := ravendb.NewSuggestionOptions()
		options.Accuracy = 0.4
		options.PageSize = 5
		options.Distance = ravendb.StringDistanceTypes_JARO_WINKLER
		options.SortMode = ravendb.SuggestionSortMode_POPULARITY

		q := session.QueryInIndexOld(ravendb.GetTypeOf(&User4{}), index)
		fn := func(x ravendb.ISuggestionBuilder) {
			x.ByField("name", "johne", "davi").WithOptions(options)
		}
		q2 := q.SuggestUsingBuilder(fn)
		suggestionQueryResult, err := q2.Execute()
		assert.NoError(t, err)

		su := suggestionQueryResult["name"].Suggestions
		assert.Equal(t, len(su), 5)
		ok := ravendb.StringArrayContainsSequence(su, []string{"john", "jones", "johnson", "david", "jack"})
		assert.True(t, ok)

		session.Close()
	}
}

func TestSuggestions(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches the order of Java tests
	suggestions_canGetSuggestions(t)
	suggestions_usingLinq_Multiple_words(t)
	suggestions_withTypo(t)
	suggestions_usingLinq(t)
	suggestions_usingLinq_WithOptions(t)
	suggestions_exactMatch(t)
}
