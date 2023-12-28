package tests

import (
	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func ravendb14989_should_work(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	{
		session := openSessionMustWithOptions(t, store, &ravendb.SessionOptions{
			Database:        "",
			RequestExecutor: nil,
			TransactionMode: ravendb.TransactionMode_ClusterWide,
			DisableAtomicDocumentWritesInClusterWideTransaction: nil,
		})

		name := "egor"
		user := &User{
			Name: &name,
		}

		clusterTransaction, err := session.Advanced().ClusterTransaction()
		assert.NoError(t, err)

		clusterTransaction.CreateCompareExchangeValue(strings.Repeat("e", 513), user)
		assert.NoError(t, err)

		err = session.SaveChanges()

		assert.Error(t, err)

		assert.True(t, strings.Contains(err.Error(), "CompareExchangeKeyTooBigException"))
	}
}

func TestRavenDB14989(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches the order of Java tests
	ravendb14989_should_work(t, driver)
}
