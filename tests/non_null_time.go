package tests

import (
	"testing"
	"time"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func TestNonNilTimeError(t *testing.T) {
	driver := createTestDriver(t)
	test_case_string_non_nil_error(t, driver)
}

func test_case_string_non_nil_error(t *testing.T, driver *RavenTestDriver) {

	id := "customer1"

	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	var time time.Time
	time, err = ravendb.ParseTime("2006-01-02T15:04:05.9999999Z")
	rtime := ravendb.Time(time)

	obj := &CustomerNilTime{
		ID:        id,
		CreatedAt: &rtime,
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

		assert.True(t, customer.CreatedAt != nil)

		session.Close()
	}
}
