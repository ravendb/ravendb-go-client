package tests

import (
	"reflect"
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

type Abc struct {
	ID string
}

type Xyz struct {
	ID string
}

func loadAllStartingWithLoadAllStartingWith(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	doc1 := &Abc{
		ID: "abc/1",
	}
	doc2 := &Xyz{
		ID: "xyz/1",
	}

	{
		session := openSessionMust(t, store)
		err = session.Store(doc1)
		assert.NoError(t, err)
		err = session.Store(doc2)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
	}

	{
		session := openSessionMust(t, store)
		args := &ravendb.StartsWithArgs{
			StartsWith: "abc/",
		}
		testClasses := session.Advanced().Lazily().LoadStartingWith(reflect.TypeOf(&Abc{}), args)

		iv, err := testClasses.GetValue()
		assert.NoError(t, err)
		v := iv.(map[string]*Abc)
		assert.Equal(t, len(v), 1)
		assert.Equal(t, v["abc/1"].ID, "abc/1")

		test2Classes := session.QueryOld(reflect.TypeOf(&Xyz{})).WaitForNonStaleResults(0).Lazily()
		iv, err = test2Classes.GetValue()
		assert.NoError(t, err)
		v2 := iv.([]*Xyz)
		assert.Equal(t, len(v2), 1)

		assert.Equal(t, v2[0].ID, "xyz/1")
	}
}

func TestLoadAllStartingWith(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	loadAllStartingWithLoadAllStartingWith(t, driver)
}
