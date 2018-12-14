package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"testing"
	"time"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func queryStreaming_canStreamQueryResults(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	index := NewUsers_ByName2()
	err = index.Execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		for i := 0; i < 200; i++ {
			err = session.Store(&User{})
			assert.NoError(t, err)
		}
		err = session.SaveChanges()
		assert.NoError(t, err)
	}
	err = driver.waitForIndexing(store, store.GetDatabase(), 0)
	assert.NoError(t, err)

	count := 0
	{
		session := openSessionMust(t, store)
		query := session.QueryInIndexOld(reflect.TypeOf(&User{}), index)
		stream, err := session.Advanced().StreamQuery(query, nil)
		assert.NoError(t, err)
		for {
			var u *User
			_, err = stream.Next(&u)
			if err != nil {
				break
			}
			count++
			assert.NotNil(t, u)
		}
		if err == io.EOF {
			err = nil
		}
		stream.Close()
		assert.NoError(t, err)
		assert.Equal(t, 200, count)
	}
}

func queryStreaming_canStreamQueryResultsWithQueryStatistics(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	index := NewUsers_ByName2()
	err = index.Execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		for i := 0; i < 100; i++ {
			err = session.Store(&User{})
			assert.NoError(t, err)
		}
		err = session.SaveChanges()
		assert.NoError(t, err)
	}
	err = driver.waitForIndexing(store, store.GetDatabase(), 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		query := session.QueryInIndexOld(reflect.TypeOf(&User{}), index)
		statsRef := &ravendb.StreamQueryStatistics{}

		stream, err := session.Advanced().StreamQuery(query, statsRef)
		assert.NoError(t, err)
		for {
			var u *User
			_, err = stream.Next(&u)
			if err != nil {
				break
			}
			assert.NotNil(t, u)
		}
		if err == io.EOF {
			err = nil
		}
		stream.Close()
		assert.NoError(t, err)

		assert.Equal(t, statsRef.IndexName, index.GetIndexName())
		assert.Equal(t, statsRef.TotalResults, 100)
		assert.Equal(t, statsRef.IndexTimestamp.Year(), time.Now().Year())
	}
}

func queryStreaming_canStreamRawQueryResults(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	index := NewUsers_ByName2()
	err = index.Execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		for i := 0; i < 200; i++ {
			err = session.Store(&User{})
			assert.NoError(t, err)
		}
		err = session.SaveChanges()
		assert.NoError(t, err)
	}
	err = driver.waitForIndexing(store, store.GetDatabase(), 0)
	assert.NoError(t, err)

	count := 0
	{
		session := openSessionMust(t, store)
		qs := fmt.Sprintf(`from index '%s'`, index.GetIndexName())
		query := session.Advanced().RawQuery(qs)
		stream, err := session.Advanced().StreamRawQuery(query, nil)
		assert.NoError(t, err)
		for {
			var u *User
			_, err = stream.Next(&u)
			if err != nil {
				break
			}
			count++
			assert.NotNil(t, u)
		}
		if err == io.EOF {
			err = nil
		}
		stream.Close()
		assert.NoError(t, err)
		assert.Equal(t, 200, count)
	}
}

func queryStreaming_canStreamRawQueryResultsWithQueryStatistics(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	index := NewUsers_ByName2()
	err = index.Execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		for i := 0; i < 100; i++ {
			err = session.Store(&User{})
			assert.NoError(t, err)
		}
		err = session.SaveChanges()
		assert.NoError(t, err)
	}
	err = driver.waitForIndexing(store, store.GetDatabase(), 0)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		statsRef := &ravendb.StreamQueryStatistics{}
		qs := fmt.Sprintf(`from index '%s'`, index.GetIndexName())
		query := session.Advanced().RawQuery(qs)
		stream, err := session.Advanced().StreamRawQuery(query, statsRef)
		assert.NoError(t, err)
		for {
			var u *User
			_, err = stream.Next(&u)
			if err != nil {
				break
			}
			assert.NotNil(t, u)
		}
		if err == io.EOF {
			err = nil
		}
		stream.Close()
		assert.NoError(t, err)

		assert.Equal(t, statsRef.IndexName, index.GetIndexName())
		assert.Equal(t, statsRef.TotalResults, 100)
		assert.Equal(t, statsRef.IndexTimestamp.Year(), time.Now().Year())
	}
}

func queryStreaming_canStreamRawQueryIntoStream(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	index := NewUsers_ByName2()
	err = index.Execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		err = session.Store(&User{})
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
	}
	err = driver.waitForIndexing(store, store.GetDatabase(), 0)
	assert.NoError(t, err)

	{
		var buf bytes.Buffer
		session := openSessionMust(t, store)
		qs := fmt.Sprintf(`from index '%s'`, index.GetIndexName())
		query := session.Advanced().RawQuery(qs)
		err = session.Advanced().StreamRawQueryInto(query, &buf)
		assert.NoError(t, err)

		var m map[string]interface{}
		err = json.Unmarshal(buf.Bytes(), &m)
		assert.NoError(t, err)
		_, ok := m["Results"]
		assert.True(t, ok)
	}
}

func queryStreaming_canStreamQueryIntoStream(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	index := NewUsers_ByName2()
	err = index.Execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		err = session.Store(&User{})
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
	}
	err = driver.waitForIndexing(store, store.GetDatabase(), 0)
	assert.NoError(t, err)

	{
		var buf bytes.Buffer
		session := openSessionMust(t, store)
		query := session.QueryInIndexOld(reflect.TypeOf(&User{}), index)
		err = session.Advanced().StreamQueryInto(query, &buf)
		assert.NoError(t, err)

		var m map[string]interface{}
		err = json.Unmarshal(buf.Bytes(), &m)
		assert.NoError(t, err)
		_, ok := m["Results"]
		assert.True(t, ok)
	}
}

// avoid conflicts with NewUsers_ByName in indexes_from_client_test.go
func NewUsers_ByName2() *ravendb.AbstractIndexCreationTask {
	res := ravendb.NewAbstractIndexCreationTask("NewUsers_ByName2")
	res.Map = "from u in docs.Users select new { u.name, lastName = u.lastName.Boost(10) }"
	res.Index("name", ravendb.FieldIndexing_SEARCH)
	res.IndexSuggestions = append(res.IndexSuggestions, "name")
	res.Store("name", ravendb.FieldStorage_YES)
	return res
}

func TestQueryStreaming(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	queryStreaming_canStreamQueryIntoStream(t, driver)
	queryStreaming_canStreamQueryResultsWithQueryStatistics(t, driver)
	queryStreaming_canStreamQueryResults(t, driver)
	queryStreaming_canStreamRawQueryResults(t, driver)
	queryStreaming_canStreamRawQueryIntoStream(t, driver)
	queryStreaming_canStreamRawQueryResultsWithQueryStatistics(t, driver)
}
