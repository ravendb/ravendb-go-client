package tests

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

func bulkInsertsTest_simpleBulkInsertShouldWork(t *testing.T) {
	fooBar1 := &FooBar{}
	fooBar1.setName("John Doe")

	fooBar2 := &FooBar{}
	fooBar2.setName("Jane Doe")

	fooBar3 := &FooBar{}
	fooBar3.setName("Mega John")

	fooBar4 := &FooBar{}
	fooBar4.setName("Mega Jane")

	var err error
	store := getDocumentStoreMust(t)
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
		doc1I, err := session.LoadOld(ravendb.GetTypeOf(&FooBar{}), "FooBars/1-A")
		assert.NoError(t, err)
		doc2I, err := session.LoadOld(ravendb.GetTypeOf(&FooBar{}), "FooBars/2-A")
		assert.NoError(t, err)
		doc3I, err := session.LoadOld(ravendb.GetTypeOf(&FooBar{}), "FooBars/3-A")
		assert.NoError(t, err)
		doc4I, err := session.LoadOld(ravendb.GetTypeOf(&FooBar{}), "FooBars/4-A")
		assert.NoError(t, err)

		assert.NotNil(t, doc1I)
		assert.NotNil(t, doc2I)
		assert.NotNil(t, doc3I)
		assert.NotNil(t, doc4I)

		doc1 := doc1I.(*FooBar)
		doc2 := doc2I.(*FooBar)
		doc3 := doc3I.(*FooBar)
		doc4 := doc4I.(*FooBar)

		assert.Equal(t, doc1.GetName(), "John Doe")
		assert.Equal(t, doc2.GetName(), "Jane Doe")
		assert.Equal(t, doc3.GetName(), "Mega John")
		assert.Equal(t, doc4.GetName(), "Mega Jane")

		session.Close()
	}
}

func bulkInsertsTest_killedToEarly(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
	}
}

func bulkInsertsTest_shouldNotAcceptIdsEndingWithPipeLine(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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

func bulkInsertsTest_canModifyMetadataWithBulkInsert(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	et := time.Now().Add(time.Hour * 24 * 365)
	expirationDate := ravendb.NetISO8601Utils_format(et)

	{
		bulkInsert := store.BulkInsert()

		fooBar := &FooBar{}
		fooBar.setName("Jon Show")
		metadata := &ravendb.MetadataAsDictionary{}
		metadata.Put(ravendb.Constants_Documents_Metadata_EXPIRES, expirationDate)

		_, err = bulkInsert.StoreWithMetadata(fooBar, metadata)
		assert.NoError(t, err)

		err = bulkInsert.Close()
		assert.NoError(t, err)
	}

	{
		session := openSessionMust(t, store)
		entity, err := session.LoadOld(ravendb.GetTypeOf(&FooBar{}), "FooBars/1-A")
		assert.NoError(t, err)

		meta, err := session.Advanced().GetMetadataFor(entity)
		assert.NoError(t, err)

		metadataExpirationDate, ok := meta.Get(ravendb.Constants_Documents_Metadata_EXPIRES)
		assert.True(t, ok)
		assert.Equal(t, metadataExpirationDate, expirationDate)
	}
}

type FooBar struct {
	Name string
}

func (f *FooBar) GetName() string {
	return f.Name
}

func (f *FooBar) setName(name string) {
	f.Name = name
}

func TestBulkInserts(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	bulkInsertsTest_simpleBulkInsertShouldWork(t)
	bulkInsertsTest_shouldNotAcceptIdsEndingWithPipeLine(t)

	// TODO: this test is flaky. Sometimes it fails as in https://travis-ci.org/kjk/ravendb-go-client/builds/404729678
	// it fails oftent if we comment out all other tests here.
	// Looks like timing issue where the server doesn't yet see the command
	// that we're trying to kill
	if ravendb.EnableFlakyTests {
		bulkInsertsTest_killedToEarly(t)
	}
	bulkInsertsTest_canModifyMetadataWithBulkInsert(t)
}
