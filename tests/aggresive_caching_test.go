package tests

import (
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
	}

	assert.Equal(t, requestExecutor.NumberOfServerRequests.Get(), 1+oldNumOfRequests)
	store.Close()
}

func aggressiveCaching_canAggressivelyCacheLoads(t *testing.T) {

}

func aggressiveCaching_canAggressivelyCacheQueries(t *testing.T) {

}

func aggressiveCaching_waitForNonStaleResultsIgnoresAggressiveCaching(t *testing.T) {

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
	// TODO: hangs in DatabaseChanges waiting for websockets end
	//aggressiveCaching_canAggressivelyCacheLoads_404(t)
}
