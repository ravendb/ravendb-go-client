package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Note: must rename as it conflicts with Order in order_test.go
type CustomerNilReference struct {
	ID        string  `json:"Id"`
	Reference *string `json:"reference"`
	Name      string  `json:"name"`
}

func TestStringNilError(t *testing.T) {
	driver := createTestDriver(t)
	test_case_string_nil_error(t, driver)
}

func test_case_string_nil_error(t *testing.T, driver *RavenTestDriver) {

	id := "customer1"
	reference := "reference"

	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	obj := &CustomerNilReference{
		ID:   id,
		Name: "customer_name",
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
		var customer *CustomerNilReference
		err = session.Load(&customer, id)

		customer.Reference = &reference

		session.SaveChanges()

		session.Close()
	}
}
