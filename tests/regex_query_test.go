package tests

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func regexQueryWueriesWithRegexFromDocumentQuery(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		err = session.Store(NewRegexMe("I love dogs and cats"))
		assert.NoError(t, err)
		err = session.Store(NewRegexMe("I love cats"))
		assert.NoError(t, err)
		err = session.Store(NewRegexMe("I love dogs"))
		assert.NoError(t, err)
		err = session.Store(NewRegexMe("I love bats"))
		assert.NoError(t, err)
		err = session.Store(NewRegexMe("dogs love me"))
		assert.NoError(t, err)
		err = session.Store(NewRegexMe("cats love me"))
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		query := session.Advanced().QueryCollectionForType(reflect.TypeOf(&RegexMe{}))
		query = query.WhereRegex("text", "^[a-z ]{2,4}love")

		var result []*RegexMe
		err = query.GetResults(&result)
		assert.NoError(t, err)
		assert.Equal(t, len(result), 4)

		iq, err := query.GetIndexQuery()
		assert.NoError(t, err)
		assert.Equal(t, iq.GetQuery(), "from RegexMes where regex(text, $p0)")
		assert.Equal(t, iq.GetQueryParameters()["p0"], "^[a-z ]{2,4}love")

		session.Close()
	}
}

type RegexMe struct {
	Text string `json:"text"`
}

func NewRegexMe(text string) *RegexMe {
	return &RegexMe{
		Text: text,
	}
}

func TestRegexQuery(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	regexQueryWueriesWithRegexFromDocumentQuery(t, driver)
}
