package ravendb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Note: client/documents/LoadTest.java

type Foo struct {
	Name string
}

func (f *Foo) getName() string {
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

func (b *Bar) getName() string {
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
		err = session.StoreEntity(foo)
		assert.NoError(t, err)

		fooId := session.advanced().getDocumentId(foo)
		bar := &Bar{}
		bar.setName("End")
		bar.setFooId(fooId)

		session.StoreEntity(bar)

		barId = session.advanced().getDocumentId(bar)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	{
		newSession := openSessionMust(t, store)
		// Note: in Java it's fooId, we must match Go naming with FooId
		bar, err := newSession.include("FooId").loadMulti(getTypeOf(&Bar{}), []string{barId})
		assert.NoError(t, err)

		assert.NotNil(t, bar)
		assert.Equal(t, len(bar), 1)
		for _, v := range bar {
			assert.NotNil(t, v)
		}

		numOfRequests := newSession.advanced().getNumberOfRequests()

		barV := bar[barId].(*Bar)
		foo, err := newSession.load(getTypeOf(&Foo{}), barV.getFooId())
		assert.NoError(t, err)
		assert.NotNil(t, foo)
		fooV := foo.(*Foo)
		assert.Equal(t, fooV.getName(), "Beginning")

		assert.Equal(t, newSession.advanced().getNumberOfRequests(), numOfRequests)
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
	defer func() {
		r := recover()
		destroyDriver()
		if r != nil {
			panic(r)
		}
	}()

	// matches order of Java tests
	documentsLoadTest_loadWithIncludes(t)
	documentsLoadTest_loadWithIncludesAndMissingDocument(t)
}
