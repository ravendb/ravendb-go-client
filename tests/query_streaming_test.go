package tests

import (
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func queryStreaming_canStreamQueryResults(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
	err = gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	assert.NoError(t, err)

	/*
	   int count = 0;

	   try (IDocumentSession session = store.openSession()) {
	       IDocumentQuery<User> query = session.query(User.class, Users_ByName2.class);
	       try (CloseableIterator<StreamResult<User>> stream = session.advanced().stream(query)) {
	           while (stream.hasNext()) {
	               StreamResult<User> user = stream.next();
	               count++;

	               assertThat(user)
	                       .isNotNull();
	           }
	       }
	   }

	   assertThat(count)
	           .isEqualTo(200);
	*/
}

func queryStreaming_canStreamQueryResultsWithQueryStatistics(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
	err = gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	assert.NoError(t, err)

	/*
	   try (IDocumentSession session = store.openSession()) {
	       IDocumentQuery<User> query = session.query(User.class, Users_ByName2.class);

	       Reference<StreamQueryStatistics> statsRef = new Reference<>();
	       try (CloseableIterator<StreamResult<User>> reader = session.advanced().stream(query, statsRef)) {
	           while (reader.hasNext()) {
	               StreamResult<User> user = reader.next();
	               assertThat(user)
	                       .isNotNull();
	           }

	           assertThat(statsRef.value.getIndexName())
	                   .isEqualTo("Users/ByName");

	           assertThat(statsRef.value.getTotalResults())
	                   .isEqualTo(100);

	           assertThat(statsRef.value.getIndexTimestamp())
	                   .isInSameYearAs(new Date());
	       }
	   }
	*/
}

func queryStreaming_canStreamRawQueryResults(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
	err = gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	assert.NoError(t, err)

	/*
	   int count = 0;

	   try (IDocumentSession session = store.openSession()) {
	       IRawDocumentQuery<User> query = session.advanced().rawQuery(User.class, "from index '" + new Users_ByName2().getIndexName() + "'");
	       try (CloseableIterator<StreamResult<User>> stream = session.advanced().stream(query)) {
	           while (stream.hasNext()) {
	               StreamResult<User> user = stream.next();
	               count++;

	               assertThat(user)
	                       .isNotNull();
	           }
	       }
	   }

	   assertThat(count)
	           .isEqualTo(200);
	*/
}

func queryStreaming_canStreamRawQueryResultsWithQueryStatistics(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
	err = gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	assert.NoError(t, err)

	/*
	   try (IDocumentSession session = store.openSession()) {
	       IRawDocumentQuery<User> query = session.advanced().rawQuery(User.class, "from index '" + new Users_ByName2().getIndexName() + "'");

	       Reference<StreamQueryStatistics> statsRef = new Reference<>();
	       try (CloseableIterator<StreamResult<User>> reader = session.advanced().stream(query, statsRef)) {
	           while (reader.hasNext()) {
	               StreamResult<User> user = reader.next();
	               assertThat(user)
	                       .isNotNull();
	           }

	           assertThat(statsRef.value.getIndexName())
	                   .isEqualTo("Users/ByName");

	           assertThat(statsRef.value.getTotalResults())
	                   .isEqualTo(100);

	           assertThat(statsRef.value.getIndexTimestamp())
	                   .isInSameYearAs(new Date());
	       }
	   }
	*/
}

func queryStreaming_canStreamRawQueryIntoStream(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
	err = gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	assert.NoError(t, err)

	/*
	   try (IDocumentSession session = store.openSession()) {
	       IRawDocumentQuery<User> query = session.advanced().rawQuery(User.class, "from index '" + new Users_ByName2().getIndexName() + "'");
	       ByteArrayOutputStream baos = new ByteArrayOutputStream();
	       session.advanced().streamInto(query, baos);

	       JsonNode queryResult = JsonExtensions.getDefaultMapper().readTree(baos.toByteArray());
	       assertThat(queryResult)
	               .isInstanceOf(ObjectNode.class);

	       assertThat(queryResult.get("Results").get(0))
	               .isNotNull();
	   }
	*/
}

func queryStreaming_canStreamQueryIntoStream(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
	err = gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	assert.NoError(t, err)

	/*
	   try (IDocumentSession session = store.openSession()) {
	       IDocumentQuery<User> query = session.query(User.class, Users_ByName2.class);
	       ByteArrayOutputStream baos = new ByteArrayOutputStream();
	       session.advanced().streamInto(query, baos);

	       JsonNode queryResult = JsonExtensions.getDefaultMapper().readTree(baos.toByteArray());
	       assertThat(queryResult)
	               .isInstanceOf(ObjectNode.class);

	       assertThat(queryResult.get("Results").get(0))
	               .isNotNull();
	   }
	*/
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

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	queryStreaming_canStreamQueryIntoStream(t)
	queryStreaming_canStreamQueryResultsWithQueryStatistics(t)
	queryStreaming_canStreamQueryResults(t)
	queryStreaming_canStreamRawQueryResults(t)
	queryStreaming_canStreamRawQueryIntoStream(t)
	queryStreaming_canStreamRawQueryResultsWithQueryStatistics(t)
}
