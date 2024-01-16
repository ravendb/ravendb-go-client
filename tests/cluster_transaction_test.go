package tests

import (
	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
	"reflect"
	"strings"
	"testing"
)

func throwOnInvalidTransactionMode(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	user := &User{}
	user.setName("Karmel")

	var session *ravendb.DocumentSession
	{
		session = openSessionMust(t, store)

		ct := session.Advanced().ClusterTransaction()
		assert.Nil(t, ct)

		session.Close()
		session = nil
	}

	disableAtomicDocumentWritesInClusterWideTransaction := true
	{
		session = openSessionMustWithOptions(t, store, &ravendb.SessionOptions{
			Database:        "",
			RequestExecutor: nil,
			TransactionMode: ravendb.TransactionMode_ClusterWide,
			DisableAtomicDocumentWritesInClusterWideTransaction: &disableAtomicDocumentWritesInClusterWideTransaction,
		})

		clusterTransaction := session.Advanced().ClusterTransaction()
		assert.NotNil(t, clusterTransaction)

		session.Advanced().ClusterTransaction().CreateCompareExchangeValue("usernames/ayende", user)
		session.Advanced().SetTransactionMode(ravendb.TransactionMode_SingleNode)

		err := session.SaveChanges()
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), "Performing cluster transaction operation require the TransactionMode to be set to TransactionMode_ClusterWide"))

		session.Advanced().SetTransactionMode(ravendb.TransactionMode_ClusterWide)

		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}
}

func testSessionSequance(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	user1 := &User{}
	user2 := &User{}

	user1.setName("Karmel")
	user2.setName("Indych")
	dat := true
	{
		session := openSessionMustWithOptions(t, store, &ravendb.SessionOptions{
			Database:        "",
			RequestExecutor: nil,
			TransactionMode: ravendb.TransactionMode_ClusterWide,
			DisableAtomicDocumentWritesInClusterWideTransaction: &dat,
		})

		var err error
		session.Advanced().ClusterTransaction().CreateCompareExchangeValue("usernames/ayende", user1)
		err = session.StoreWithID(user1, "users/1")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		value, err := session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(&User{}), "usernames/ayende")
		value.Value = user2

		session.StoreWithID(user2, "users/2")
		user1.setAge(10)
		session.StoreWithID(user1, "users/1")
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}
}

func testSessionOnPrimitiveType(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()
	dat := true
	options := &ravendb.SessionOptions{
		Database:        "",
		RequestExecutor: nil,
		TransactionMode: ravendb.TransactionMode_ClusterWide,
		DisableAtomicDocumentWritesInClusterWideTransaction: &dat,
	}

	{
		session := openSessionMustWithOptions(t, store, options)
		defer session.Close()

		var err error
		_, err = session.Advanced().ClusterTransaction().CreateCompareExchangeValue("int/Key", 1)
		assert.NoError(t, err)

		_, err = session.Advanced().ClusterTransaction().CreateCompareExchangeValue("string/Key", "hello")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		value, err := session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(int(0)), "int/Key")
		assert.NoError(t, err)
		assert.Equal(t, 1, value.Value)

		value, err = session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(string("")), "string/Key")
		assert.NoError(t, err)
		assert.Equal(t, "hello", value.Value)
	}

	{
		session := openSessionMustWithOptions(t, store, options)
		defer session.Close()

		value, err := session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(int(0)), "int/Key")
		assert.NoError(t, err)
		assert.Equal(t, 1, value.Value)
		value.Value = 2

		value, err = session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(int(0)), "int/Key")
		assert.NoError(t, err)
		assert.Equal(t, 2, value.Value)

		value, err = session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(string("")), "string/Key")
		assert.NoError(t, err)
		assert.Equal(t, "hello", value.Value)
		value.Value = "world"

		value, err = session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(string("")), "string/Key")
		assert.NoError(t, err)
		assert.Equal(t, "world", value.Value)

		err = session.SaveChanges()
		assert.NoError(t, err)
	}

	{
		session := openSessionMustWithOptions(t, store, options)
		defer session.Close()

		value, err := session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(int(0)), "int/Key")
		assert.NoError(t, err)
		assert.Equal(t, 2, value.Value)

		value, err = session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(string("")), "string/Key")
		assert.NoError(t, err)
		assert.Equal(t, "world", value.Value)
	}

	{
		session := openSessionMustWithOptions(t, store, options)
		defer session.Close()

		value, err := session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(int(0)), "int/Key")
		assert.NoError(t, err)
		err = session.Advanced().ClusterTransaction().DeleteCompareExchangeValueByKey("int/Key", value.GetIndex())

		value, err = session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(string("")), "string/Key")
		assert.NoError(t, err)
		err = session.Advanced().ClusterTransaction().DeleteCompareExchangeValue(value)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)
	}

	{
		session := openSessionMustWithOptions(t, store, options)
		defer session.Close()

		value, err := session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(int(0)), "int/Key")
		assert.NoError(t, err)
		assert.Nil(t, value)

		value, err = session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(string("")), "string/Key")
		assert.NoError(t, err)
		assert.Nil(t, value)
	}
}

func canCreateClusterTransactionRequest(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	user1 := &User{}
	user3 := &User{}

	user1.setName("Karmel")
	user3.setName("Indych")
	dat := true
	{
		session := openSessionMustWithOptions(t, store, &ravendb.SessionOptions{
			Database:        "",
			RequestExecutor: nil,
			TransactionMode: ravendb.TransactionMode_ClusterWide,
			DisableAtomicDocumentWritesInClusterWideTransaction: &dat,
		})

		var err error
		_, err = session.Advanced().ClusterTransaction().CreateCompareExchangeValue("usernames/ayende", user1)
		assert.NoError(t, err)

		err = session.StoreWithID(user3, "foo/bar")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		cev, err := session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(&User{}), "usernames/ayende")
		assert.NoError(t, err)

		userFromCEV, ok := cev.Value.(*User)
		assert.True(t, ok)

		user := &User{}
		err = session.Load(&user, "foo/bar")
		assert.NoError(t, err)

		assert.Equal(t, user1.Name, userFromCEV.Name)
		assert.Equal(t, user.Name, user3.Name)
		session.Close()
	}
}

func canDeleteCompareExchangeValue(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	user1 := &User{}
	user3 := &User{}

	user1.setName("Karmel")
	user3.setName("Indych")

	opt := true
	options := &ravendb.SessionOptions{
		Database:        "",
		RequestExecutor: nil,
		TransactionMode: ravendb.TransactionMode_ClusterWide,
		DisableAtomicDocumentWritesInClusterWideTransaction: &opt,
	}

	{
		session := openSessionMustWithOptions(t, store, options)
		defer session.Close()
		session.Advanced().ClusterTransaction().CreateCompareExchangeValue("usernames/ayende", user1)
		session.Advanced().ClusterTransaction().CreateCompareExchangeValue("usernames/marcin", user3)
		session.SaveChanges()
	}

	{
		session := openSessionMustWithOptions(t, store, options)
		defer session.Close()

		compareExchangeValue, err := session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(&User{}), "usernames/ayende")
		assert.NoError(t, err)
		assert.NotNil(t, compareExchangeValue)

		err = session.Advanced().ClusterTransaction().DeleteCompareExchangeValue(compareExchangeValue)
		assert.NoError(t, err)

		compareExchangeValue2, err := session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(&User{}), "usernames/marcin")
		assert.NoError(t, err)
		assert.NotNil(t, compareExchangeValue2)
		err = session.Advanced().ClusterTransaction().DeleteCompareExchangeValueByKey(compareExchangeValue2.GetKey(), compareExchangeValue2.GetIndex())
		assert.NoError(t, err)
		session.SaveChanges()
	}

	{
		session := openSessionMustWithOptions(t, store, options)
		defer session.Close()

		compareExchangeValue, err := session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(&User{}), "usernames/ayende")
		assert.NoError(t, err)
		assert.Nil(t, compareExchangeValue)

		compareExchangeValue, err = session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(&User{}), "usernames/marcin")
		assert.NoError(t, err)
		assert.Nil(t, compareExchangeValue)
	}

}

func TestClusterTransaction(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)
	throwOnInvalidTransactionMode(t, driver)
	testSessionSequance(t, driver)
	canCreateClusterTransactionRequest(t, driver)
	canDeleteCompareExchangeValue(t, driver)
	testSessionOnPrimitiveType(t, driver)
}
