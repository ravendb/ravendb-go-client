package tests

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func loadIntoStreamCanLoadByIdsIntoStream(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	insertData(t, store)

	{
		session := openSessionMust(t, store)

		stream := bytes.NewBuffer(nil)

		ids := []string{"employee2s/1-A", "employee2s/4-A", "employee2s/7-A"}
		err = session.Advanced().LoadIntoStream(ids, stream)
		assert.NoError(t, err)

		d, err := ioutil.ReadAll(stream)
		assert.NoError(t, err)
		var jsonNode map[string]interface{}
		err = json.Unmarshal(d, &jsonNode)
		assert.NoError(t, err)

		res := jsonNode["Results"]
		a := res.([]interface{})
		assert.Equal(t, len(a), 3)

		names := []string{"Aviv", "Maxim", "Michael"}

		for _, v := range a {
			v2 := v.(ravendb.ObjectNode)
			s, _ := ravendb.JsonGetAsText(v2, "firstName")
			assert.True(t, stringArrayContains(names, s))
		}

		session.Close()
	}
}

func loadIntoStreamCanLoadStartingWithIntoStream(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	insertData(t, store)

	{
		session := openSessionMust(t, store)
		stream := bytes.NewBuffer(nil)

		args := &ravendb.StartsWithArgs{
			StartsWith: "employee2s/",
		}
		err = session.Advanced().LoadStartingWithIntoStream(stream, args)
		assert.NoError(t, err)

		d, err := ioutil.ReadAll(stream)
		assert.NoError(t, err)
		var jsonNode map[string]interface{}
		err = json.Unmarshal(d, &jsonNode)
		assert.NoError(t, err)

		res := jsonNode["Results"]
		a := res.([]interface{})
		assert.Equal(t, len(a), 7)

		names := []string{"Aviv", "Iftah", "Tal", "Maxim", "Karmel", "Grisha", "Michael"}
		for _, v := range a {
			v2 := v.(ravendb.ObjectNode)
			s, _ := ravendb.JsonGetAsText(v2, "firstName")
			assert.True(t, stringArrayContains(names, s))
		}

		session.Close()
	}
}

func insertData(t *testing.T, store *ravendb.IDocumentStore) {
	var err error
	{
		session := openSessionMust(t, store)

		insertEmployee := func(name string) error {
			employee := &Employee2{
				FirstName: name,
			}
			return session.Store(employee)
		}

		err = insertEmployee("Aviv")
		assert.NoError(t, err)
		err = insertEmployee("Iftah")
		assert.NoError(t, err)
		err = insertEmployee("Tal")
		assert.NoError(t, err)
		err = insertEmployee("Maxim")
		assert.NoError(t, err)
		err = insertEmployee("Karmel")
		assert.NoError(t, err)
		err = insertEmployee("Grisha")
		assert.NoError(t, err)
		err = insertEmployee("Michael")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}
}

// Note: conflicts with Employee in employee_test.go
type Employee2 struct {
	FirstName string `json:"firstName"`
}

func TestLoadIntoStream(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	loadIntoStreamCanLoadStartingWithIntoStream(t, driver)
	loadIntoStreamCanLoadByIdsIntoStream(t, driver)
}
