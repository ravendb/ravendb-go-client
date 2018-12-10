package tests

import (
	"reflect"
	"testing"
	"time"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func initAggressiveCaching(t *testing.T) *ravendb.DocumentStore {
	var err error
	store := getDocumentStoreMust(t)
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

func aggressiveCaching_canAggressivelyCacheLoads_404(t *testing.T) {
	defer disableLogFailedRequests()()

	store := initAggressiveCaching(t)
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

func aggressiveCaching_canAggressivelyCacheLoads(t *testing.T) {
	store := initAggressiveCaching(t)
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

func aggressiveCaching_canAggressivelyCacheQueries(t *testing.T) {
	store := initAggressiveCaching(t)
	requestExecutor := store.GetRequestExecutor()

	oldNumOfRequests := requestExecutor.NumberOfServerRequests.Get()
	for i := 0; i < 5; i++ {
		session := openSessionMust(t, store)
		{
			dur := time.Minute * 5
			context := session.Advanced().GetDocumentStore().AggressivelyCacheFor(dur)
			q := session.QueryOld(reflect.TypeOf(&User{}))
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

func aggressiveCaching_waitForNonStaleResultsIgnoresAggressiveCaching(t *testing.T) {
	store := initAggressiveCaching(t)
	requestExecutor := store.GetRequestExecutor()

	oldNumOfRequests := requestExecutor.NumberOfServerRequests.Get()
	for i := 0; i < 5; i++ {
		session := openSessionMust(t, store)
		{
			dur := time.Minute * 5
			context := session.Advanced().GetDocumentStore().AggressivelyCacheFor(dur)
			q := session.QueryOld(reflect.TypeOf(&User{}))
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
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	aggressiveCaching_canAggressivelyCacheQueries(t)
	aggressiveCaching_waitForNonStaleResultsIgnoresAggressiveCaching(t)
	aggressiveCaching_canAggressivelyCacheLoads(t)
	aggressiveCaching_canAggressivelyCacheLoads_404(t)
}
