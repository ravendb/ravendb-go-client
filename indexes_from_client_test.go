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
	defer store.Close()

	{
		session := openSessionMust(t, store)
		user1 := NewUser()
		user1.setName("Marcin")
		err = session.StoreEntityWithID(user1, "users/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
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
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	indexes := []*AbstractIndexCreationTask{NewUsersIndex()}
	err = store.executeIndexes(indexes)
	assert.NoError(t, err)

	indexNamesOperation := NewGetIndexNamesOperation(0, 10)
	err = store.maintenance().send(indexNamesOperation)
	assert.NoError(t, err)
	indexNames := indexNamesOperation.Command.Result
	assert.Equal(t, len(indexNames), 1)
}

func indexesFromClientTest_canDelete(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	userIndex := NewUsersIndex()
	err = store.executeIndex(userIndex)
	assert.NoError(t, err)

	op := NewDeleteIndexOperation(NewUsersIndex().getIndexName())
	err = store.maintenance().send(op)
	assert.NoError(t, err)

	command := NewGetStatisticsCommand()
	err = store.getRequestExecutor().executeCommand(command)
	assert.NoError(t, err)
	statistics := command.Result
	assert.Equal(t, len(statistics.getIndexes()), 0)
}

func indexesFromClientTest_canStopAndStart(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	err = NewUsers_ByName().execute(store)
	assert.NoError(t, err)

	{
		op := NewGetIndexingStatusOperation()
		err = store.maintenance().send(op)
		assert.NoError(t, err)
		status := op.Command.Result

		assert.Equal(t, status.getStatus(), IndexRunningStatus_RUNNING)
		assert.Equal(t, len(status.getIndexes()), 1)
		assert.Equal(t, status.getIndexes()[0].getStatus(), IndexRunningStatus_RUNNING)
	}

	{
		op := NewStopIndexingOperation()
		err = store.maintenance().send(op)
		assert.NoError(t, err)

		{
			op := NewGetIndexingStatusOperation()
			err = store.maintenance().send(op)
			assert.NoError(t, err)
			status := op.Command.Result
			assert.Equal(t, status.getStatus(), IndexRunningStatus_PAUSED)
			assert.Equal(t, status.getIndexes()[0].getStatus(), IndexRunningStatus_PAUSED)
		}
	}

	indexName := ""
	{
		op := NewStartIndexingOperation()
		err = store.maintenance().send(op)
		assert.NoError(t, err)
		{
			op := NewGetIndexingStatusOperation()
			err = store.maintenance().send(op)
			assert.NoError(t, err)
			status := op.Command.Result
			indexName = status.getIndexes()[0].getName()

			assert.Equal(t, status.getStatus(), IndexRunningStatus_RUNNING)
			assert.Equal(t, len(status.getIndexes()), 1)
			assert.Equal(t, status.getIndexes()[0].getStatus(), IndexRunningStatus_RUNNING)
		}

	}

	{
		op := NewStopIndexOperation(indexName)
		err = store.maintenance().send(op)
		assert.NoError(t, err)
		{
			op := NewGetIndexingStatusOperation()
			err = store.maintenance().send(op)
			assert.NoError(t, err)
			status := op.Command.Result
			assert.Equal(t, status.getStatus(), IndexRunningStatus_RUNNING)
			assert.Equal(t, len(status.getIndexes()), 1)
			assert.Equal(t, status.getIndexes()[0].getStatus(), IndexRunningStatus_PAUSED)
		}
	}
}

func indexesFromClientTest_setLockModeAndSetPriority(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()
	{
		session := openSessionMust(t, store)

		user1 := NewUser()
		user1.setName("Fitzchak")
		err = session.StoreEntity(user1)
		assert.NoError(t, err)

		user2 := NewUser()
		user2.setName("Arek")
		err = session.StoreEntity(user2)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)
		q := session.query(getTypeOf(&User{}))
		q = q.waitForNonStaleResults(0)
		// TODO: should this be Name (name of the struct field) and we would
		// convert that to json tag (if necessary) internally?
		q = q.whereEquals("name", "Arek")
		users, err := q.toList()
		assert.NoError(t, err)
		assert.Equal(t, len(users), 1)
	}

	op := NewGetIndexesOperation(0, 128)
	err = store.maintenance().send(op)
	assert.NoError(t, err)
	indexes := op.Command.Result
	assert.Equal(t, len(indexes), 1)

	index := indexes[0]

	{
		op := NewGetIndexStatisticsOperation(index.getName())
		err = store.maintenance().send(op)
		assert.NoError(t, err)
		stats := op.Command.Result
		assert.Equal(t, stats.getLockMode(), IndexLockMode_UNLOCK)
		assert.Equal(t, stats.getPriority(), IndexPriority_NORMAL)
	}

	{
		op := NewSetIndexesLockOperation(index.getName(), IndexLockMode_LOCKED_IGNORE)
		err = store.maintenance().send(op)
		assert.NoError(t, err)
	}

	{
		op := NewSetIndexesPriorityOperation(index.getName(), IndexPriority_LOW)
		err = store.maintenance().send(op)
		assert.NoError(t, err)
	}
	{
		op := NewGetIndexStatisticsOperation(index.getName())
		err = store.maintenance().send(op)
		assert.NoError(t, err)
		stats := op.Command.Result
		assert.Equal(t, stats.getLockMode(), IndexLockMode_LOCKED_IGNORE)
		assert.Equal(t, stats.getPriority(), IndexPriority_LOW)
	}
}

func indexesFromClientTest_getTerms(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user1 := NewUser()
		user1.setName("Fitzchak")
		err = session.StoreEntity(user1)
		assert.NoError(t, err)

		user2 := NewUser()
		user2.setName("Arek")
		err = session.StoreEntity(user2)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	var indexName string

	{
		session := openSessionMust(t, store)

		var stats *QueryStatistics
		q := session.query(getTypeOf(&User{}))
		q = q.waitForNonStaleResults(0)
		q = q.statistics(&stats)
		q = q.whereEquals("name", "Arek")
		_, err := q.toList()
		assert.NoError(t, err)

		indexName = stats.getIndexName()

		session.Close()
	}

	op := NewGetTermsOperationWithPageSize(indexName, "name", "", 128)
	err = store.maintenance().send(op)
	assert.NoError(t, err)
	terms := op.Command.Result
	assert.Equal(t, len(terms), 2)
	assert.True(t, stringArrayContains(terms, "fitzchak"))
	assert.True(t, stringArrayContains(terms, "arek"))
}

func indexesFromClientTest_getIndexNames(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user1 := NewUser()
		user1.setName("Fitzchak")
		err = session.StoreEntity(user1)
		assert.NoError(t, err)

		user2 := NewUser()
		user2.setName("Arek")
		err = session.StoreEntity(user2)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	var indexName string

	{
		session := openSessionMust(t, store)

		var stats *QueryStatistics
		q := session.query(getTypeOf(&User{}))
		q = q.waitForNonStaleResults(0)
		q = q.statistics(&stats)
		q = q.whereEquals("name", "Arek")
		_, err := q.toList()
		assert.NoError(t, err)

		indexName = stats.getIndexName()

		session.Close()
	}

	{
		session := openSessionMust(t, store)
		op := NewGetIndexNamesOperation(0, 10)
		err = store.maintenance().send(op)
		assert.NoError(t, err)

		indexNames := op.Command.Result

		assert.Equal(t, len(indexNames), 1)
		assert.True(t, stringArrayContains(indexNames, indexName))
		session.Close()
	}
}

func indexesFromClientTest_canExplain(t *testing.T) {
	// TODO: requires query
}
func indexesFromClientTest_moreLikeThis(t *testing.T) {
	// TODO: requires query
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
