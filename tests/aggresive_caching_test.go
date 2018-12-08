package tests

import (
	"testing"
)

func aggressiveCaching_canAggressivelyCacheQueries(t *testing.T) {

}

func aggressiveCaching_waitForNonStaleResultsIgnoresAggressiveCaching(t *testing.T) {

}

func aggressiveCaching_canAggressivelyCacheLoads(t *testing.T) {

}

func aggressiveCaching_canAggressivelyCacheLoads_404(t *testing.T) {

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
