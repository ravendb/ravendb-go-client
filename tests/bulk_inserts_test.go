package tests

import (
	"strings"
	"testing"
	"time"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func bulkInsertsTest_simpleBulkInsertShouldWork(t *testing.T, driver *RavenTestDriver) {
	fooBar1 := &FooBar{}
	fooBar1.Name = "John Doe"

	fooBar2 := &FooBar{}
	fooBar2.Name = "Jane Doe"

	fooBar3 := &FooBar{}
	fooBar3.Name = "Mega John"

	fooBar4 := &FooBar{}
	fooBar4.Name = "Mega Jane"

	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		bulkInsert := store.BulkInsert()

		_, err = bulkInsert.Store(fooBar1)
		assert.NoError(t, err)

		_, err = bulkInsert.Store(fooBar2)
		assert.NoError(t, err)

		_, err = bulkInsert.Store(fooBar3)
		assert.NoError(t, err)

		_, err = bulkInsert.Store(fooBar4)
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

func bulkInsertsTest_killedToEarly(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		bulkInsert := store.BulkInsert()

		_, err = bulkInsert.Store(&FooBar{})
		assert.NoError(t, err)
		err = bulkInsert.Abort()
		assert.NoError(t, err)
		_, err = bulkInsert.Store(&FooBar{})
		assert.Error(t, err)

		_, ok := err.(*ravendb.BulkInsertAbortedException)
		assert.True(t, ok)

		err = bulkInsert.Close()
		assert.NoError(t, err)
	}
}

func bulkInsertsTest_shouldNotAcceptIdsEndingWithPipeLine(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		bulkInsert := store.BulkInsert()
		err = bulkInsert.StoreWithID(&FooBar{}, "foobars|", nil)
		assert.Error(t, err)
		_, ok := err.(*ravendb.UnsupportedOperationException)
		assert.True(t, ok)
		ok = strings.Contains(err.Error(), "Document ids cannot end with '|', but was called with foobars|")
		assert.True(t, ok)

		err = bulkInsert.Close()
		assert.NoError(t, err)
	}
}

func bulkInsertsTest_canModifyMetadataWithBulkInsert(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	et := time.Now().Add(time.Hour * 24 * 365)
	expirationDate := ravendb.NetISO8601Utils_format(et)

	{
		bulkInsert := store.BulkInsert()

		fooBar := &FooBar{}
		fooBar.Name = "Jon Show"
		metadata := &ravendb.MetadataAsDictionary{}
		metadata.Put(ravendb.Constants_Documents_Metadata_EXPIRES, expirationDate)

		_, err = bulkInsert.StoreWithMetadata(fooBar, metadata)
		assert.NoError(t, err)

		err = bulkInsert.Close()
		assert.NoError(t, err)
	}

	{
		session := openSessionMust(t, store)
		var entity *FooBar
		err = session.Load(&entity, "FooBars/1-A")
		assert.NoError(t, err)

		// TODO: should this be GetMetadataFor(entity)? Should we support both?
		meta, err := session.Advanced().GetMetadataFor(&entity)
		assert.NoError(t, err)

		metadataExpirationDate, ok := meta.Get(ravendb.Constants_Documents_Metadata_EXPIRES)
		assert.True(t, ok)
		assert.Equal(t, metadataExpirationDate, expirationDate)
	}
}

type FooBar struct {
	Name string
}

func TestBulkInserts(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	bulkInsertsTest_simpleBulkInsertShouldWork(t, driver)
	bulkInsertsTest_shouldNotAcceptIdsEndingWithPipeLine(t, driver)

	// TODO: this test is flaky. Sometimes it fails as in https://travis-ci.org/kjk/ravendb-go-client/builds/404729678
	// it fails oftent if we comment out all other tests here.
	// Looks like timing issue where the server doesn't yet see the command
	// that we're trying to kill
	if ravendb.EnableFlakyTests {
		bulkInsertsTest_killedToEarly(t, driver)
	}
	bulkInsertsTest_canModifyMetadataWithBulkInsert(t, driver)
}
