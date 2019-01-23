package tests

import (
	"strings"
	"testing"
	"time"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func bulkInsertsTestSimpleBulkInsertShouldWork(t *testing.T, driver *RavenTestDriver) {
	fooBar1 := &FooBar{}
	fooBar1.Name = "John Doe"

	fooBar2 := &FooBar{}
	fooBar2.Name = "Jane Doe"

	fooBar3 := &FooBar{}
	fooBar3.Name = "Mega John"

	fooBar4 := &FooBar{}
	fooBar4.Name = "Mega Jane"

	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		bulkInsert := store.BulkInsert()

		_, err = bulkInsert.Store(fooBar1, nil)
		assert.NoError(t, err)

		_, err = bulkInsert.Store(fooBar2, nil)
		assert.NoError(t, err)

		_, err = bulkInsert.Store(fooBar3, nil)
		assert.NoError(t, err)

		_, err = bulkInsert.Store(fooBar4, nil)
		assert.NoError(t, err)

		err = bulkInsert.Close()
		assert.NoError(t, err)
	}

	{
		session := openSessionMust(t, store)
		var doc1, doc2, doc3, doc4 *FooBar
		err = session.Load(&doc1, "FooBars/1-A")
		assert.NoError(t, err)
		err = session.Load(&doc2, "FooBars/2-A")
		assert.NoError(t, err)
		err = session.Load(&doc3, "FooBars/3-A")
		assert.NoError(t, err)
		err = session.Load(&doc4, "FooBars/4-A")
		assert.NoError(t, err)

		assert.Equal(t, doc1.Name, "John Doe")
		assert.Equal(t, doc2.Name, "Jane Doe")
		assert.Equal(t, doc3.Name, "Mega John")
		assert.Equal(t, doc4.Name, "Mega Jane")

		session.Close()
	}
}

func bulkInsertsTestKilledToEarly(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		bulkInsert := store.BulkInsert()

		_, err = bulkInsert.Store(&FooBar{}, nil)
		assert.NoError(t, err)
		err = bulkInsert.Abort()
		if err == nil {
			_, err = bulkInsert.Store(&FooBar{}, nil)
		}
		if err == nil {
			err = bulkInsert.Close()
		}

		assert.Error(t, err)
		if enableFlakyTests {
			// TODO: this fails always on windows and occasionally on Linux on CI
			_, ok := err.(*ravendb.BulkInsertAbortedError)
			assert.True(t, ok, "expected error to be of type ravendb.BulkInsertAbortedError, got type '%T', value: '%s'", err, err)
		}
	}
}

func bulkInsertsTestShouldNotAcceptIdsEndingWithPipeLine(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		bulkInsert := store.BulkInsert()
		err = bulkInsert.StoreWithID(&FooBar{}, "foobars|", nil)
		assert.Error(t, err)
		_, ok := err.(*ravendb.UnsupportedOperationError)
		assert.True(t, ok)
		ok = strings.Contains(err.Error(), "Document ids cannot end with '|', but was called with foobars|")
		assert.True(t, ok)

		err = bulkInsert.Close()
		assert.NoError(t, err)
	}
}

func bulkInsertsTestCanModifyMetadataWithBulkInsert(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	et := time.Now().Add(time.Hour * 24 * 365)
	expirationDate := ravendb.Time(et).Format()

	{
		bulkInsert := store.BulkInsert()

		fooBar := &FooBar{}
		fooBar.Name = "Jon Show"
		metadata := &ravendb.MetadataAsDictionary{}
		metadata.Put(ravendb.MetadataExpires, expirationDate)

		_, err = bulkInsert.Store(fooBar, metadata)
		assert.NoError(t, err)

		err = bulkInsert.Close()
		assert.NoError(t, err)
	}

	{
		session := openSessionMust(t, store)
		var entity *FooBar
		err = session.Load(&entity, "FooBars/1-A")
		assert.NoError(t, err)

		meta, err := session.Advanced().GetMetadataFor(entity)
		assert.NoError(t, err)

		metadataExpirationDate, ok := meta.Get(ravendb.MetadataExpires)
		assert.True(t, ok)
		assert.Equal(t, metadataExpirationDate, expirationDate)
	}
}

type FooBar struct {
	Name string
}

func TestBulkInserts(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	bulkInsertsTestSimpleBulkInsertShouldWork(t, driver)
	bulkInsertsTestShouldNotAcceptIdsEndingWithPipeLine(t, driver)
	bulkInsertsTestKilledToEarly(t, driver)
	bulkInsertsTestCanModifyMetadataWithBulkInsert(t, driver)
}
