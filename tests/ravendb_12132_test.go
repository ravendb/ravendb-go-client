package tests

import (
	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func canCreateClusterTransactionRequest1(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	{
		session := openSessionMustWithOptions(t, store, &ravendb.SessionOptions{
			Database:        "",
			RequestExecutor: nil,
			TransactionMode: ravendb.TransactionMode_ClusterWide,
			DisableAtomicDocumentWritesInClusterWideTransaction: nil,
		})

		user := &Document{ID: "this/is/my/id", Name: "Grisha"}
		clusterTransaction, err := session.Advanced().ClusterTransaction()
		assert.NoError(t, err)

		_, err = clusterTransaction.CreateCompareExchangeValue("usernames/ayende", user)

		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		var result *ravendb.CompareExchangeValue
		clusterTransaction, err = session.Advanced().ClusterTransaction()

		result, err = clusterTransaction.GetCompareExchangeValue(reflect.TypeOf(&Document{}), "usernames/ayende")
		userFromCluster, cast := result.GetValue().(*Document)
		assert.True(t, cast)
		assert.NoError(t, err)
		assert.Equal(t, user.Name, userFromCluster.Name)
		assert.Equal(t, user.ID, userFromCluster.ID)

		session.Close()
	}

	{
		operation, err := ravendb.NewGetCompareExchangeValueOperation(reflect.TypeOf(&Document{}), "usernames/ayende")
		assert.NoError(t, err)

		err = store.Operations().Send(operation, nil)
		assert.NoError(t, err)

		result := operation.Command.Result

		assert.NotNil(t, result)

		docFromServer := result.Value.(*Document)
		assert.Equal(t, "Grisha", docFromServer.Name)
		assert.Equal(t, "this/is/my/id", docFromServer.ID)
	}

	store.Close()
}

func TestRavenDB12132(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches the order of Java tests
	canCreateClusterTransactionRequest1(t, driver)
}
