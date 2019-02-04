package tests

import (
	"sync"
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func ravendb10566_shouldBeAvailable(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	var name string
	var mu sync.Mutex

	afterSaveChanges := func(event *ravendb.AfterSaveChangesEventArgs) {
		meta := event.GetDocumentMetadata()

		nameI, ok := meta.Get("Name")
		assert.True(t, ok)

		mu.Lock()
		defer mu.Unlock()
		name = nameI.(string)
	}
	store.AddAfterSaveChangesListener(afterSaveChanges)

	{
		session := openSessionMust(t, store)

		user := &User{}
		user.setName("Oren")

		err = session.StoreWithID(user, "users/oren")
		assert.NoError(t, err)

		metadata, err := session.Advanced().GetMetadataFor(user)
		assert.NoError(t, err)
		metadata.Put("Name", "FooBar")

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	assert.Equal(t, name, "FooBar")
}

func TestRavenDB10566(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches Java's order
	ravendb10566_shouldBeAvailable(t, driver)
}
