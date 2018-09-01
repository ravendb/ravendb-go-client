package tests

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func regexQuery_queriesWithRegexFromDocumentQuery(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
		query := session.Advanced().DocumentQueryOld(reflect.TypeOf(&RegexMe{}))
		query = query.WhereRegex("text", "^[a-z ]{2,4}love")

		iq := query.GetIndexQuery()
		assert.Equal(t, iq.GetQuery(), "from RegexMes where regex(text, $p0)")

		assert.Equal(t, iq.GetQueryParameters()["p0"], "^[a-z ]{2,4}love")

		result, err := query.ToListOld()
		assert.NoError(t, err)
		assert.Equal(t, len(result), 4)

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

func (r *RegexMe) getText() string {
	return r.Text
}

func (r *RegexMe) setText(text string) {
	r.Text = text
}

func TestRegexQuery(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	regexQuery_queriesWithRegexFromDocumentQuery(t)
}
