package tests

import (
	"testing"
	"time"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func initAggressiveCaching(t *testing.T, driver *RavenTestDriver) *ravendb.DocumentStore {
	var err error
	store := driver.getDocumentStoreMust(t)
	store.DisableAggressiveCaching()

	{
		session := openSessionMust(t, store)
		err = session.Store(&User{})
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	return store
}

func aggressiveCachingCanAggressivelyCacheLoads404(t *testing.T, driver *RavenTestDriver) {
	defer disableLogFailedRequests()()

	store := initAggressiveCaching(t, driver)
	requestExecutor := store.GetRequestExecutor()

	oldNumOfRequests := requestExecutor.NumberOfServerRequests.Get()
	for i := 0; i < 5; i++ {
		session := openSessionMust(t, store)
		{
			dur := time.Minute * 5
			context := session.Advanced().GetDocumentStore().AggressivelyCacheFor(dur)
			session.Load(&User{}, "users/not-there")
			context.Close()
		}
		session.Close()
	}

	currNo := requestExecutor.NumberOfServerRequests.Get()

	assert.Equal(t, currNo, 1+oldNumOfRequests)
	store.Close()
}

func aggressiveCachingCanAggressivelyCacheLoads(t *testing.T, driver *RavenTestDriver) {
	store := initAggressiveCaching(t, driver)
	requestExecutor := store.GetRequestExecutor()

	oldNumOfRequests := requestExecutor.NumberOfServerRequests.Get()
	for i := 0; i < 5; i++ {
		session := openSessionMust(t, store)
		{
			dur := time.Minute * 5
			context := session.Advanced().GetDocumentStore().AggressivelyCacheFor(dur)
			var u *User
			session.Load(&u, "users/1-A")
			context.Close()
		}
		session.Close()
	}
	currNo := requestExecutor.NumberOfServerRequests.Get()
	assert.Equal(t, currNo, 1+oldNumOfRequests)
}

func aggressiveCachingCanAggressivelyCacheQueries(t *testing.T, driver *RavenTestDriver) {
	store := initAggressiveCaching(t, driver)
	requestExecutor := store.GetRequestExecutor()

	oldNumOfRequests := requestExecutor.NumberOfServerRequests.Get()
	for i := 0; i < 5; i++ {
		session := openSessionMust(t, store)
		{
			dur := time.Minute * 5
			context := session.Advanced().GetDocumentStore().AggressivelyCacheFor(dur)
			q := session.Query()
			var u []*User
			err := q.ToList(&u)
			assert.NoError(t, err)
			context.Close()
		}
		session.Close()
	}
	currNo := requestExecutor.NumberOfServerRequests.Get()
	assert.Equal(t, currNo, 1+oldNumOfRequests)
}

func aggressiveCachingWaitForNonStaleResultsIgnoresAggressiveCaching(t *testing.T, driver *RavenTestDriver) {
	store := initAggressiveCaching(t, driver)
	requestExecutor := store.GetRequestExecutor()

	oldNumOfRequests := requestExecutor.NumberOfServerRequests.Get()
	for i := 0; i < 5; i++ {
		session := openSessionMust(t, store)
		{
			dur := time.Minute * 5
			context := session.Advanced().GetDocumentStore().AggressivelyCacheFor(dur)
			q := session.Query()
			q = q.WaitForNonStaleResults(0)
			var u []*User
			err := q.ToList(&u)
			assert.NoError(t, err)
			context.Close()
		}
		session.Close()
	}
	currNo := requestExecutor.NumberOfServerRequests.Get()
	assert.NotEqual(t, currNo, 1+oldNumOfRequests)
}

func TestAggressiveCaching(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	aggressiveCachingCanAggressivelyCacheQueries(t, driver)
	aggressiveCachingWaitForNonStaleResultsIgnoresAggressiveCaching(t, driver)
	aggressiveCachingCanAggressivelyCacheLoads(t, driver)
	aggressiveCachingCanAggressivelyCacheLoads404(t, driver)
}
