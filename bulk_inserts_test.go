package ravendb

import (
	"strings"
	"testing"
	"time"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
	"github.com/stretchr/testify/assert"
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
		bulkInsert := store.bulkInsert()

		_, err = bulkInsert.store(fooBar1)
		assert.NoError(t, err)

		_, err = bulkInsert.store(fooBar2)
		assert.NoError(t, err)

		_, err = bulkInsert.store(fooBar3)
		assert.NoError(t, err)

		_, err = bulkInsert.store(fooBar4)
		assert.NoError(t, err)

		err = bulkInsert.Close()
		assert.NoError(t, err)
	}

	{
		session := openSessionMust(t, store)
		doc1I, err := session.load(getTypeOf(&FooBar{}), "FooBars/1-A")
		assert.NoError(t, err)
		doc2I, err := session.load(getTypeOf(&FooBar{}), "FooBars/2-A")
		assert.NoError(t, err)
		doc3I, err := session.load(getTypeOf(&FooBar{}), "FooBars/3-A")
		assert.NoError(t, err)
		doc4I, err := session.load(getTypeOf(&FooBar{}), "FooBars/4-A")
		assert.NoError(t, err)

		assert.NotNil(t, doc1I)
		assert.NotNil(t, doc2I)
		assert.NotNil(t, doc3I)
		assert.NotNil(t, doc4I)

		doc1 := doc1I.(*FooBar)
		doc2 := doc2I.(*FooBar)
		doc3 := doc3I.(*FooBar)
		doc4 := doc4I.(*FooBar)

		assert.Equal(t, doc1.getName(), "John Doe")
		assert.Equal(t, doc2.getName(), "Jane Doe")
		assert.Equal(t, doc3.getName(), "Mega John")
		assert.Equal(t, doc4.getName(), "Mega Jane")

		session.Close()
	}
}

func bulkInsertsTest_killedToEarly(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		bulkInsert := store.bulkInsert()

		_, err = bulkInsert.store(&FooBar{})
		assert.NoError(t, err)
		err = bulkInsert.abort()
		assert.NoError(t, err)
		time.Sleep(time.Second)
		_, err = bulkInsert.store(&FooBar{})
		assert.Error(t, err)

		_, ok := err.(*BulkInsertAbortedException)
		assert.True(t, ok)

		err = bulkInsert.Close()
		assert.NoError(t, err)
	}
}

func bulkInsertsTest_shouldNotAcceptIdsEndingWithPipeLine(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		bulkInsert := store.bulkInsert()
		err = bulkInsert.storeWithID(&FooBar{}, "foobars|", nil)
		assert.Error(t, err)
		_, ok := err.(*UnsupportedOperationException)
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
	expirationDate := NetISO8601Utils_format(et)

	{
		bulkInsert := store.bulkInsert()

		fooBar := &FooBar{}
		fooBar.setName("Jon Show")
		metadata := &MetadataAsDictionary{}
		metadata.put(Constants_Documents_Metadata_EXPIRES, expirationDate)

		_, err = bulkInsert.storeWithMetadata(fooBar, metadata)
		assert.NoError(t, err)

		err = bulkInsert.Close()
		assert.NoError(t, err)
	}

	{
		session := openSessionMust(t, store)
		entity, err := session.load(getTypeOf(&FooBar{}), "FooBars/1-A")
		assert.NoError(t, err)

		meta, err := session.advanced().getMetadataFor(entity)
		assert.NoError(t, err)

		metadataExpirationDate, ok := meta.get(Constants_Documents_Metadata_EXPIRES)
		assert.True(t, ok)
		assert.Equal(t, metadataExpirationDate, expirationDate)
	}
}

type FooBar struct {
	Name string
}

func (f *FooBar) getName() string {
	return f.Name
}

func (f *FooBar) setName(name string) {
	f.Name = name
}

func TestBulkInserts(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	if useProxy() {
		proxy.ChangeLogFile("trace_bulk_inserts_go.txt")
	}

	if false {
		dumpFailedHTTP = true
	}

	createTestDriver()
	defer deleteTestDriver()

	// matches order of Java tests
	bulkInsertsTest_simpleBulkInsertShouldWork(t)
	bulkInsertsTest_shouldNotAcceptIdsEndingWithPipeLine(t)

	// TODO: those still fail
	//bulkInsertsTest_killedToEarly(t)
	bulkInsertsTest_canModifyMetadataWithBulkInsert(t)
}
