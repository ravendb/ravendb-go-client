package tests

import (
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

// Note: must rename as it conflicts with Order in order_test.go
type CustomerNilTime struct {
	ID        string        `json:"Id"`
	CreatedAt *ravendb.Time `json:"createdAt"`
}

func TestNilTimeError(t *testing.T) {
	driver := createTestDriver(t)
	test_case_string_nil_error(t, driver)
}

func test_case_nil_time_error(t *testing.T, driver *RavenTestDriver) {

	id := "customer1"

	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	obj := &CustomerNilTime{
		ID:        id,
		CreatedAt: nil,
	}

	{
		session := openSessionMust(t, store)
		err = session.Store(obj)
		assert.NoError(t, err)
		session.SaveChanges()
		session.Close()
	}

	{
		session := openSessionMust(t, store)
		var customer *CustomerNilTime
		err = session.Load(&customer, id)

		assert.True(t, customer.CreatedAt == nil)

		session.Close()
	}
}
