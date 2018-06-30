package ravendb

import (
	"testing"
	"time"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

func NewUsersIndex() *AbstractIndexCreationTask {
	res := NewAbstractIndexCreationTask("UsersIndex")
	res.smap = "from user in docs.users select new { user.name }"
	return res
}

func NewUsers_ByName() *AbstractIndexCreationTask {
	res := NewAbstractIndexCreationTask("NewUsers_ByName")
	res.smap = "from u in docs.Users select new { u.name }"

	res.index("name", FieldIndexing_SEARCH)

	res.indexSuggestions.add("name")

	res.store("name", FieldStorage_YES)

	return res
}

func Posts_ByTitleAndDesc() *AbstractIndexCreationTask {
	res := NewAbstractIndexCreationTask("Posts_ByTitleAndDesc")
	res.smap = "from p in docs.Posts select new { p.title, p.desc }"
	res.index("title", FieldIndexing_SEARCH)
	res.store("title", FieldStorage_YES)
	res.analyze("title", "Lucene.Net.Analysis.SimpleAnalyzer")

	res.index("desc", FieldIndexing_SEARCH)
	res.store("desc", FieldStorage_YES)
	res.analyze("desc", "Lucene.Net.Analysis.SimpleAnalyzer")

	return res
}

func indexesFromClientTest_canReset(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)

	{
		session := openSessionMust(t, store)
		user1 := NewUser()
		user1.setName("Marcin")
		err = session.StoreEntityWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
	}

	userIndex := NewUsersIndex()
	err = store.executeIndex(userIndex)
	assert.NoError(t, err)

	err = gRavenTestDriver.waitForIndexing(store, store.getDatabase(), 0)
	assert.NoError(t, err)

	command := NewGetStatisticsCommand()
	err = store.getRequestExecutor().executeCommand(command)
	assert.NoError(t, err)
	statistics := command.Result
	firstIndexingTime := statistics.getIndexes()[0].getLastIndexingTime()

	indexName := NewUsersIndex().getIndexName()
	// now reset index
	time.Sleep(time.Millisecond * 2)
	{
		op := NewResetIndexOperation(indexName)
		err = store.maintenance().send(op)
		assert.NoError(t, err)
	}

	err = gRavenTestDriver.waitForIndexing(store, store.getDatabase(), 0)
	assert.NoError(t, err)

	command = NewGetStatisticsCommand()
	err = store.getRequestExecutor().executeCommand(command)
	assert.NoError(t, err)
	statistics = command.Result
	secondIndexingTime := statistics.getLastIndexingTime()
	assert.True(t, secondIndexingTime.Sub(firstIndexingTime) > 0)
}

func indexesFromClientTest_canExecuteManyIndexes(t *testing.T) {
}
func indexesFromClientTest_canDelete(t *testing.T) {
}
func indexesFromClientTest_canStopAndStart(t *testing.T) {
}
func indexesFromClientTest_setLockModeAndSetPriority(t *testing.T) {
}
func indexesFromClientTest_getTerms(t *testing.T) {
}
func indexesFromClientTest_getIndexNames(t *testing.T) {
}
func indexesFromClientTest_canExplain(t *testing.T) {
}
func indexesFromClientTest_moreLikeThis(t *testing.T) {
}

func TestIndexesFromClient(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	if useProxy() {
		proxy.ChangeLogFile("trace_indexes_from_client_go.txt")
	}

	createTestDriver()
	defer deleteTestDriver()

	verboseLog = true
	// order matches Java tests
	indexesFromClientTest_canExecuteManyIndexes(t)
	indexesFromClientTest_canDelete(t)
	indexesFromClientTest_canReset(t)
	indexesFromClientTest_getIndexNames(t)
	indexesFromClientTest_canStopAndStart(t)
	indexesFromClientTest_canExplain(t)
	indexesFromClientTest_moreLikeThis(t)
	indexesFromClientTest_setLockModeAndSetPriority(t)
	indexesFromClientTest_getTerms(t)
}
