package tests

import (
	"testing"

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

func documentsLoadTest_loadWithIncludes(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
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
		bars := map[string]*Bar{}
		err = newSession.Include("FooId").LoadMulti(bars, []string{barId})
		assert.NoError(t, err)

		assert.NotNil(t, bars)
		assert.Equal(t, len(bars), 1)
		for _, v := range bars {
			assert.NotNil(t, v)
		}

		numOfRequests := newSession.Advanced().GetNumberOfRequests()

		bar := bars[barId]
		var foo *Foo
		err = newSession.Load(&foo, bar.FooId)
		assert.NoError(t, err)
		assert.NotNil(t, foo)
		assert.Equal(t, foo.Name, "Beginning")

		assert.Equal(t, newSession.Advanced().GetNumberOfRequests(), numOfRequests)
		newSession.Close()
	}
}

func documentsLoadTest_loadWithIncludesAndMissingDocument(t *testing.T, driver *RavenTestDriver) {
	// TODO: is @Disabled
}

func TestDocumentsLoad(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	documentsLoadTest_loadWithIncludes(t, driver)
	documentsLoadTest_loadWithIncludesAndMissingDocument(t, driver)
}
