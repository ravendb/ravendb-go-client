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
	store := driver.getDocumentStoreMust(t)
	options := &ravendb.SessionOptions{
		Database:        "",
		RequestExecutor: nil,
		TransactionMode: ravendb.TransactionMode_ClusterWide,
	}
	var err error
	{
		session := openSessionMustWithOptions(t, store, options)

		company := &Company2{Id: "companies/1", ExternalId: "companies/cf", Name: "CF"}
		session.Store(company)

		numberOfRequest := session.Advanced().GetNumberOfRequests()
		address := &Address{City: "Torun"}

		assert.NotNil(t, session.Advanced().ClusterTransaction())

		_, err = session.Advanced().ClusterTransaction().CreateCompareExchangeValue(company.ExternalId, address)
		assert.NoError(t, err)

		assert.Equal(t, numberOfRequest, session.Advanced().GetNumberOfRequests())

		value1, err := session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(&Address{}), company.ExternalId)
		assert.NoError(t, err)
		assert.Equal(t, numberOfRequest, session.Advanced().GetNumberOfRequests())

		assert.Equal(t, address, value1.GetValue())
		assert.Equal(t, company.ExternalId, value1.GetKey())
		assert.Equal(t, int64(0), value1.GetIndex())
		err = session.SaveChanges()
		assert.NoError(t, err)

		assert.Equal(t, numberOfRequest+1, session.Advanced().GetNumberOfRequests())

		value1, err = session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(&Address{}), company.ExternalId)
		assert.NoError(t, err)
		assert.Equal(t, address, value1.GetValue())
		assert.Equal(t, company.ExternalId, value1.GetKey())
		assert.True(t, value1.GetIndex() > int64(0))

		value2, err := session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(&Address{}), company.ExternalId)
		assert.NoError(t, err)
		assert.Equal(t, numberOfRequest+1, session.Advanced().GetNumberOfRequests())

		assert.Equal(t, value1, value2)

		session.SaveChanges()
		assert.Equal(t, numberOfRequest+1, session.Advanced().GetNumberOfRequests())

		session.Advanced().Clear()
		value3, err := session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(&Address{}), company.ExternalId)
		assert.NoError(t, err)
		assert.Equal(t, value3, value2)

		session.Close()
	}

	{
		session := openSessionMustWithOptions(t, store, options)
		address := &Address{City: "Hadera"}

		assert.NotNil(t, session.Advanced().ClusterTransaction())

		session.Advanced().ClusterTransaction().CreateCompareExchangeValue("companies/hr", address)
		session.SaveChanges()
		session.Close()
	}

	{
		session := openSessionMustWithOptions(t, store, options)
		numberOfRequest := session.Advanced().GetNumberOfRequests()

		assert.NotNil(t, session.Advanced().ClusterTransaction())

		value1, err := session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(&Address{}), "companies/cf")
		assert.NoError(t, err)
		assert.Equal(t, numberOfRequest+1, session.Advanced().GetNumberOfRequests())

		value2, err := session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(&Address{}), "companies/hr")
		assert.NoError(t, err)
		assert.Equal(t, numberOfRequest+2, session.Advanced().GetNumberOfRequests())

		values, err := session.Advanced().ClusterTransaction().GetCompareExchangeValuesWithKeys(reflect.TypeOf(&Address{}), []string{"companies/cf", "companies/hr"})
		assert.Equal(t, numberOfRequest+2, session.Advanced().GetNumberOfRequests())
		assert.Equal(t, 2, len(values))
		assert.Equal(t, value1, values[value1.GetKey()])
		assert.Equal(t, value2, values[value2.GetKey()])

		values, err = session.Advanced().ClusterTransaction().GetCompareExchangeValuesWithKeys(reflect.TypeOf(&Address{}), []string{"companies/cf", "companies/hr", "companies/hx"})
		assert.Equal(t, 3, len(values))
		assert.Equal(t, numberOfRequest+3, session.Advanced().GetNumberOfRequests())
		assert.Equal(t, value1, values[value1.GetKey()])
		assert.Equal(t, value2, values[value2.GetKey()])

		value3, err := session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(&Address{}), "companies/hx")
		assert.NoError(t, err)
		assert.Equal(t, numberOfRequest+3, session.Advanced().GetNumberOfRequests())

		assert.Nil(t, value3)
		assert.Nil(t, values["companies/hx"])
		err = session.SaveChanges()
		assert.Equal(t, numberOfRequest+3, session.Advanced().GetNumberOfRequests())

		assert.NoError(t, err)

		address := &Address{City: "Bydgoszcz"}
		_, err = session.Advanced().ClusterTransaction().CreateCompareExchangeValue("companies/hx", address)
		assert.NoError(t, err)
		session.SaveChanges()
		assert.Equal(t, numberOfRequest+4, session.Advanced().GetNumberOfRequests())

		session.Close()
	}
}

func compareExchangeValueTrackingInSession_NoTracking(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	options := &ravendb.SessionOptions{
		Database:        "",
		RequestExecutor: nil,
		TransactionMode: ravendb.TransactionMode_ClusterWide,
	}
	company := &Company2{Id: "companies/1", ExternalId: "companies/cf", Name: "CF"}
	{
		session := openSessionMustWithOptions(t, store, options)
		session.NoTracking(true)

		session.Store(company)

		address := &Address{City: "Torun"}

		assert.NotNil(t, session.Advanced().ClusterTransaction())

		session.Advanced().ClusterTransaction().CreateCompareExchangeValue(company.ExternalId, address)
		session.SaveChanges()
		session.Close()
	}

	{
		session := openSessionMustWithOptions(t, store, options)
		session.NoTracking(true)
		defer session.Close()

		numberOfRequests := session.Advanced().GetNumberOfRequests()
		assert.NotNil(t, session.Advanced().ClusterTransaction())

		value1, err := session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(&Address{}), company.ExternalId)
		assert.NoError(t, err)
		assert.Equal(t, numberOfRequests+1, session.Advanced().GetNumberOfRequests())

		value2, err := session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(&Address{}), company.ExternalId)
		assert.NoError(t, err)

		assert.Equal(t, numberOfRequests+2, session.Advanced().GetNumberOfRequests())
		assert.True(t, value1 != value2)

		value3, err := session.Advanced().ClusterTransaction().GetCompareExchangeValue(reflect.TypeOf(&Address{}), company.ExternalId)
		assert.NoError(t, err)
		assert.Equal(t, numberOfRequests+3, session.Advanced().GetNumberOfRequests())
		assert.True(t, value2 != value3)
	}

	{
		session := openSessionMustWithOptions(t, store, options)
		session.NoTracking(true)
		defer session.Close()

		numberOfRequests := session.Advanced().GetNumberOfRequests()
		assert.NotNil(t, session.Advanced().ClusterTransaction())

		value1, err := session.Advanced().ClusterTransaction().GetCompareExchangeValues(reflect.TypeOf(&Address{}), company.ExternalId, 0, 25)
		assert.NoError(t, err)
		assert.Equal(t, numberOfRequests+1, session.Advanced().GetNumberOfRequests())

		value2, err := session.Advanced().ClusterTransaction().GetCompareExchangeValues(reflect.TypeOf(&Address{}), company.ExternalId, 0, 25)
		assert.NoError(t, err)
		assert.Equal(t, numberOfRequests+2, session.Advanced().GetNumberOfRequests())

		assert.False(t, mapEquals(value1, value2))

		value3, err := session.Advanced().ClusterTransaction().GetCompareExchangeValues(reflect.TypeOf(&Address{}), company.ExternalId, 0, 25)
		assert.NoError(t, err)
		assert.Equal(t, numberOfRequests+3, session.Advanced().GetNumberOfRequests())
		assert.False(t, mapEquals(value3, value2))
	}

	{
		session := openSessionMustWithOptions(t, store, options)
		session.NoTracking(true)
		defer session.Close()

		numberOfRequests := session.Advanced().GetNumberOfRequests()
		assert.NotNil(t, session.Advanced().ClusterTransaction())

		value1, err := session.Advanced().ClusterTransaction().GetCompareExchangeValuesWithKeys(reflect.TypeOf(&Address{}), []string{company.ExternalId})
		assert.NoError(t, err)
		assert.Equal(t, numberOfRequests+1, session.Advanced().GetNumberOfRequests())

		value2, err := session.Advanced().ClusterTransaction().GetCompareExchangeValuesWithKeys(reflect.TypeOf(&Address{}), []string{company.ExternalId})
		assert.NoError(t, err)
		assert.Equal(t, numberOfRequests+2, session.Advanced().GetNumberOfRequests())

		assert.False(t, mapEquals(value1, value2))

		value3, err := session.Advanced().ClusterTransaction().GetCompareExchangeValuesWithKeys(reflect.TypeOf(&Address{}), []string{company.ExternalId})
		assert.NoError(t, err)
		assert.Equal(t, numberOfRequests+3, session.Advanced().GetNumberOfRequests())
		assert.False(t, mapEquals(value2, value3))
	}
}

// Compare values by pointers
func mapEquals(a, b map[string]*ravendb.CompareExchangeValue) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		vb, exist := b[k]
		if exist == false {
			return false
		}

		if v != vb {
			return false
		}
	}

	return true
}

func Test_ravendb_14006(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	compareExchangeValueTrackingInSession(t, driver)
	compareExchangeValueTrackingInSession_NoTracking(t, driver)
}
