package tests

import (
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

// Note: client/documents/LoadTest.java

type Foo struct {
	Name string
}

type Bar struct {
	FooId  string
	FooIDs []string
	Name   string
}

func documentsLoadTest_loadWithIncludes(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	barId := ""
	{
		session := openSessionMust(t, store)
		foo := &Foo{}
		foo.Name = "Beginning"
		err = session.Store(foo)
		assert.NoError(t, err)

		fooId := session.Advanced().GetDocumentID(foo)
		bar := &Bar{}
		bar.Name = "End"
		bar.FooId = fooId

		session.Store(bar)

		barId = session.Advanced().GetDocumentID(bar)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		newSession := openSessionMust(t, store)
		// Note: in Java it's fooId, we must match Go naming with FooId
		bar, err := newSession.Include("FooId").LoadMultiOld(ravendb.GetTypeOf(&Bar{}), []string{barId})
		assert.NoError(t, err)

		assert.NotNil(t, bar)
		assert.Equal(t, len(bar), 1)
		for _, v := range bar {
			assert.NotNil(t, v)
		}

		numOfRequests := newSession.Advanced().GetNumberOfRequests()

		barV := bar[barId].(*Bar)
		var foo *Foo
		err = newSession.Load(&foo, barV.FooId)
		assert.NoError(t, err)
		assert.NotNil(t, foo)
		assert.Equal(t, foo.Name, "Beginning")

		assert.Equal(t, newSession.Advanced().GetNumberOfRequests(), numOfRequests)
		newSession.Close()
	}
}

func documentsLoadTest_loadWithIncludesAndMissingDocument(t *testing.T) {
	// TODO: is @Disabled
}

func TestDocumentsLoad(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	documentsLoadTest_loadWithIncludes(t)
	documentsLoadTest_loadWithIncludesAndMissingDocument(t)
}
