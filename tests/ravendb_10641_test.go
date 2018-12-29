package tests

import (
	"testing"

	"github.com/ravendb/ravendb-go-client"

	"github.com/stretchr/testify/assert"
)

func ravendb_10641_canEditObjectsInMetadata(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		v := &Document{}
		err = session.StoreWithID(v, "items/first")
		assert.NoError(t, err)

		items := map[string]string{
			"lang": "en",
		}
		var meta *ravendb.MetadataAsDictionary
		meta, err = session.Advanced().GetMetadataFor(v)
		assert.NoError(t, err)
		meta.Put("Items", items)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var v *Document
		err = session.Load(&v, "items/first")
		assert.NoError(t, err)

		var m *ravendb.MetadataAsDictionary
		m, err = session.Advanced().GetMetadataFor(&v)
		assert.NoError(t, err)
		metadataI, ok := m.Get("Items")
		assert.True(t, ok)
		metadata := metadataI.(map[string]interface{})
		metadata["lang"] = "sv"
		// Note: unlike Java we can't intercept modifications so we have to
		// manually mark as dirty
		m.MarkDirty()

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var v *Document
		err = session.Load(&v, "items/first")
		assert.NoError(t, err)
		var metadata *ravendb.MetadataAsDictionary
		metadata, err = session.Advanced().GetMetadataFor(&v)
		assert.NoError(t, err)
		metadata.Put("test", "123")

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var v *Document
		err = session.Load(&v, "items/first")
		assert.NoError(t, err)
		_, err = session.Advanced().GetMetadataFor(&v)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var v *Document
		err = session.Load(&v, "items/first")
		assert.NoError(t, err)
		metadata, err := session.Advanced().GetMetadataFor(&v)
		assert.NoError(t, err)
		mI, ok := metadata.Get("Items")
		assert.True(t, ok)
		m := mI.(map[string]interface{})
		val := m["lang"]
		assert.Equal(t, val, "sv")

		val, ok = metadata.Get("test")
		assert.True(t, ok)
		assert.Equal(t, val, "123")

		session.Close()
	}
}

type Document struct {
	ID string
}

func TestRavenDB10641(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches the order of Java tests
	ravendb_10641_canEditObjectsInMetadata(t, driver)
}
