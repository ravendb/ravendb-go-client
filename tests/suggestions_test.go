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
	indexDefinition.Maps = ravendb.NewStringSetFromStrings("from doc in docs.Users select new { doc.name }")
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

		q := session.QueryWithQuery(ravendb.GetTypeOf(&User4{}), ravendb.Query_index("test"))
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

func suggestions_usingLinq(t *testing.T)                {}
func suggestions_usingLinq_WithOptions(t *testing.T)    {}
func suggestions_usingLinq_Multiple_words(t *testing.T) {}
func suggestions_withTypo(t *testing.T)                 {}
func suggestions_canGetSuggestions(t *testing.T)        {}

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
