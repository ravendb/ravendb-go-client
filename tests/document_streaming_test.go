package tests

import (
	"io"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func documentStreamingCanStreamDocumentsStartingWith(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		for i := 0; i < 200; i++ {
			user := &User{}
			err = session.Store(user)
			assert.NoError(t, err)
		}
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	count := 0
	{
		session := openSessionMust(t, store)
		{
			args := &ravendb.StartsWithArgs{
				StartsWith: "users/",
			}
			reader, err := session.Advanced().Stream(args)
			assert.NoError(t, err)
			for {
				var user *User
				_, err = reader.Next(&user)
				if err == io.EOF {
					err = nil
					break
				}
				assert.NoError(t, err)
				assert.NotNil(t, user)
				count++
			}
			assert.NoError(t, err)
		}
	}
	assert.Equal(t, count, 200)
}

func documentStreamingStreamWithoutIterationDoesntLeakConnection(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		for i := 0; i < 200; i++ {
			user := &User{}
			err = session.Store(user)
			assert.NoError(t, err)
		}
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	for i := 0; i < 5; i++ {
		session := openSessionMust(t, store)
		args := &ravendb.StartsWithArgs{
			StartsWith: "users/",
		}
		reader, err := session.Advanced().Stream(args)
		assert.NoError(t, err)
		// don't iterate
		reader.Close()
	}
}

func TestDocumentStreaming(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	documentStreamingCanStreamDocumentsStartingWith(t, driver)
	documentStreamingStreamWithoutIterationDoesntLeakConnection(t, driver)
}
