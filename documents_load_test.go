package ravendb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Note: client/documents/LoadTest.java

type Foo struct {
	Name string
}

func (f *Foo) GetName() string {
	return f.Name
}

func (f *Foo) setName(name string) {
	f.Name = name
}

type Bar struct {
	FooId  string
	FooIDs []string
	Name   string
}

func (b *Bar) getFooId() string {
	return b.FooId
}

func (b *Bar) setFooId(fooId string) {
	b.FooId = fooId
}

func (b *Bar) getFooIDs() []string {
	return b.FooIDs
}

func (b *Bar) setFooIDs(fooIDs []string) {
	b.FooIDs = fooIDs
}

func (b *Bar) GetName() string {
	return b.Name
}

func (b *Bar) setName(name string) {
	b.Name = name
}

func documentsLoadTest_loadWithIncludes(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	barId := ""
	{
		session := openSessionMust(t, store)
		foo := &Foo{}
		foo.setName("Beginning")
		err = session.Store(foo)
		assert.NoError(t, err)

		fooId := session.Advanced().GetDocumentID(foo)
		bar := &Bar{}
		bar.setName("End")
		bar.setFooId(fooId)

		session.Store(bar)

		barId = session.Advanced().GetDocumentID(bar)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		newSession := openSessionMust(t, store)
		// Note: in Java it's fooId, we must match Go naming with FooId
		bar, err := newSession.Include("FooId").loadMulti(GetTypeOf(&Bar{}), []string{barId})
		assert.NoError(t, err)

		assert.NotNil(t, bar)
		assert.Equal(t, len(bar), 1)
		for _, v := range bar {
			assert.NotNil(t, v)
		}

		numOfRequests := newSession.Advanced().GetNumberOfRequests()

		barV := bar[barId].(*Bar)
		foo, err := newSession.Load(GetTypeOf(&Foo{}), barV.getFooId())
		assert.NoError(t, err)
		assert.NotNil(t, foo)
		fooV := foo.(*Foo)
		assert.Equal(t, fooV.GetName(), "Beginning")

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
