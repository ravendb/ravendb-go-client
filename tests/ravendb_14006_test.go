package tests

import (
	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

type Company2 struct {
	Id         string  `json:"Id"`
	ExternalId string  `json:"ExternalId"`
	Name       string  `json:"Name"`
	Address    Address `json:"Address"`
}

func compareExchangeValueTrackingInSession(t *testing.T, driver *RavenTestDriver) {
	//ravendb.WithFiddler()

	store := driver.getDocumentStoreMust(t)
	options := &ravendb.SessionOptions{
		Database:        "",
		RequestExecutor: nil,
		TransactionMode: ravendb.TransactionMode_ClusterWide,
	}

	{
		session := openSessionMustWithOptions(t, store, options)

		company := &Company2{Id: "companies/1", ExternalId: "companies/cf", Name: "CF"}
		session.Store(company)

		numberOfRequest := session.Advanced().GetNumberOfRequests()
		address := &Address{City: "Torun"}

		clusterTransaction, err := session.Advanced().ClusterTransaction()
		assert.NoError(t, err)

		_, err = clusterTransaction.CreateCompareExchangeValue(company.ExternalId, address)
		assert.NoError(t, err)

		assert.Equal(t, numberOfRequest, session.Advanced().GetNumberOfRequests())

		value1, err := clusterTransaction.GetCompareExchangeValue(reflect.TypeOf(&Address{}), company.ExternalId)
		assert.NoError(t, err)
		assert.Equal(t, numberOfRequest, session.Advanced().GetNumberOfRequests())

		assert.Equal(t, address, value1.GetValue())
		assert.Equal(t, company.ExternalId, value1.GetKey())
		assert.Equal(t, int64(0), value1.GetIndex())

		err = session.SaveChanges()
		assert.NoError(t, err)

		assert.Equal(t, numberOfRequest+1, session.Advanced().GetNumberOfRequests())

		assert.Equal(t, address, value1.GetValue())
		assert.Equal(t, company.ExternalId, value1.GetKey())
		assert.True(t, value1.GetIndex() > int64(0))

		value2, err := clusterTransaction.GetCompareExchangeValue(reflect.TypeOf(&Address{}), company.ExternalId)
		assert.NoError(t, err)
		assert.Equal(t, numberOfRequest+1, session.Advanced().GetNumberOfRequests())

		assert.Equal(t, value1, value2)

		session.SaveChanges()
		assert.Equal(t, numberOfRequest+1, session.Advanced().GetNumberOfRequests())

		session.Advanced().Clear()
		value3, err := clusterTransaction.GetCompareExchangeValue(reflect.TypeOf(&Address{}), company.ExternalId)
		assert.NoError(t, err)
		assert.Equal(t, value3, value2)

		session.Close()
	}

	{
		session := openSessionMustWithOptions(t, store, options)
		address := &Address{City: "Hadera"}

		ct, err := session.Advanced().ClusterTransaction()
		assert.NoError(t, err)

		ct.CreateCompareExchangeValue("companies/hr", address)
		session.SaveChanges()
		session.Close()
	}

	{
		session := openSessionMustWithOptions(t, store, options)
		numberOfRequest := session.Advanced().GetNumberOfRequests()

		ct, err := session.Advanced().ClusterTransaction()
		assert.NoError(t, err)

		value1, err := ct.GetCompareExchangeValue(reflect.TypeOf(&Address{}), "companies/cf")
		assert.NoError(t, err)
		assert.Equal(t, numberOfRequest+1, session.Advanced().GetNumberOfRequests())

		value2, err := ct.GetCompareExchangeValue(reflect.TypeOf(&Address{}), "companies/hr")
		assert.NoError(t, err)
		assert.Equal(t, numberOfRequest+2, session.Advanced().GetNumberOfRequests())

		values, err := ct.GetCompareExchangeValues(reflect.TypeOf(&Address{}), []string{"companies/cf", "companies/hr"})
		assert.Equal(t, numberOfRequest+2, session.Advanced().GetNumberOfRequests())
		assert.Equal(t, 2, len(values))
		assert.Equal(t, value1, values[value1.GetKey()])
		assert.Equal(t, value2, values[value2.GetKey()])

		values, err = ct.GetCompareExchangeValues(reflect.TypeOf(&Address{}), []string{"companies/cf", "companies/hr", "companies/hx"})
		assert.Equal(t, 3, len(values))
		assert.Equal(t, numberOfRequest+3, session.Advanced().GetNumberOfRequests())
		assert.Equal(t, value1, values[value1.GetKey()])
		assert.Equal(t, value2, values[value2.GetKey()])

		value3, err := ct.GetCompareExchangeValue(reflect.TypeOf(&Address{}), "companies/hx")
		assert.NoError(t, err)
		assert.Equal(t, numberOfRequest+3, session.Advanced().GetNumberOfRequests())

		assert.Nil(t, value3)
		assert.Nil(t, values["companies/hx"])
		err = session.SaveChanges()
		assert.NoError(t, err)

		address := &Address{City: "Bydgoszcz"}
		_, err = ct.CreateCompareExchangeValue("companies/hx", address)
		assert.NoError(t, err)
		session.SaveChanges()
		assert.Equal(t, numberOfRequest+4, session.Advanced().GetNumberOfRequests())

		session.Close()
	}
}

func Test_ravendb_14006(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	compareExchangeValueTrackingInSession(t, driver)
}
