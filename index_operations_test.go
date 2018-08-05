package ravendb

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func NewUsersInvalidIndex() *AbstractIndexCreationTask {
	res := NewAbstractIndexCreationTask("UsersInvalidIndex")
	res.smap = "from u in docs.Users select new { a = 5 / u.Age }"
	return res
}

func NewUsers_Index() *AbstractIndexCreationTask {
	res := NewAbstractIndexCreationTask("Users_Index")
	res.smap = "from u in docs.Users select new { u.name }"
	return res
}

func testIndexCanDeleteIndex(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	err = index.execute(store)
	assert.NoError(t, err)

	op := NewGetIndexNamesOperation(0, 10)
	err = store.maintenance().send(op)
	assert.NoError(t, err)
	indexNames := op.Command.Result
	assert.True(t, stringArrayContains(indexNames, "UsersIndex"))

	op2 := NewDeleteIndexOperation("UsersIndex")
	err = store.maintenance().send(op2)
	assert.NoError(t, err)

	op3 := NewGetIndexNamesOperation(0, 10)
	err = store.maintenance().send(op3)
	assert.NoError(t, err)
	indexNames = op3.Command.Result
	assert.Equal(t, len(indexNames), 0)
}

func testIndexCanDisableAndEnableIndex(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	err = index.execute(store)
	assert.NoError(t, err)

	{
		op := NewDisableIndexOperation("UsersIndex")
		err = store.maintenance().send(op)
		assert.NoError(t, err)
	}

	{
		op := NewGetIndexingStatusOperation()
		err = store.maintenance().send(op)
		assert.NoError(t, err)
		indexingStatus := op.Command.Result
		indexStatus := indexingStatus.getIndexes()[0]
		assert.Equal(t, indexStatus.getStatus(), IndexRunningStatus_DISABLED)
	}

	{
		op := NewEnableIndexOperation("UsersIndex")
		err = store.maintenance().send(op)
		assert.NoError(t, err)
	}

	{
		op := NewGetIndexingStatusOperation()
		err = store.maintenance().send(op)
		assert.NoError(t, err)
		indexingStatus := op.Command.Result
		indexStatus := indexingStatus.getIndexes()[0]
		assert.Equal(t, indexStatus.getStatus(), IndexRunningStatus_RUNNING)
	}
}

func testIndexGetCanIndexes(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	err = index.execute(store)
	assert.NoError(t, err)

	{
		op := NewGetIndexesOperation(0, 10)
		err = store.maintenance().send(op)
		assert.NoError(t, err)
		indexDefinitions := op.Command.Result
		assert.Equal(t, len(indexDefinitions), 1)
	}
}

func testIndexGetCanIndexesStats(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	index := NewUsersIndex()
	err = index.execute(store)
	assert.NoError(t, err)

	{
		op := NewGetIndexesStatisticsOperation()
		err = store.maintenance().send(op)
		assert.NoError(t, err)
		indexStats := op.Command.Result
		assert.Equal(t, len(indexStats), 1)
	}
}

func testIndexGetTerms(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	err = index.execute(store)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		user := NewUser()
		user.setName("Marcin")
		err = session.StoreEntity(user)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	err = gRavenTestDriver.waitForIndexing(store, store.getDatabase(), 0)
	assert.NoError(t, err)

	{
		op := NewGetTermsOperation("UsersIndex", "name", "")
		err = store.maintenance().send(op)
		assert.NoError(t, err)
		terms := op.Command.Result
		assert.Equal(t, len(terms), 1)
		assert.Equal(t, terms[0], "marcin")
	}
}

func testIndexHasIndexChanged(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	indexDef := index.createIndexDefinition()
	op := NewPutIndexesOperation(indexDef)
	err = store.maintenance().send(op)
	assert.NoError(t, err)
	op2 := NewIndexHasChangedOperation(indexDef)
	err = store.maintenance().send(op2)
	assert.NoError(t, err)
	{
		cmd := op2.Command
		assert.False(t, cmd.Result)
	}
	m := NewStringSetFromStrings("from users")
	indexDef.setMaps(m)

	op3 := NewIndexHasChangedOperation(indexDef)
	err = store.maintenance().send(op3)
	assert.NoError(t, err)
	{
		cmd := op3.Command
		assert.True(t, cmd.Result)
	}
}

func testIndexCanStopStartIndexing(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	indexDef := index.createIndexDefinition()
	{
		op := NewPutIndexesOperation(indexDef)
		err = store.maintenance().send(op)
		assert.NoError(t, err)
	}

	{
		op := NewStopIndexingOperation()
		err = store.maintenance().send(op)
		assert.NoError(t, err)
	}

	{
		op := NewGetIndexingStatusOperation()
		err = store.maintenance().send(op)
		assert.NoError(t, err)
		indexingStatus := op.Command.Result
		indexStatus := indexingStatus.getIndexes()[0]
		assert.Equal(t, indexStatus.getStatus(), IndexRunningStatus_PAUSED)
	}

	{
		op := NewStartIndexingOperation()
		err = store.maintenance().send(op)
		assert.NoError(t, err)
	}

	{
		op := NewGetIndexingStatusOperation()
		err = store.maintenance().send(op)
		assert.NoError(t, err)
		indexingStatus := op.Command.Result
		indexStatus := indexingStatus.getIndexes()[0]
		assert.Equal(t, indexStatus.getStatus(), IndexRunningStatus_RUNNING)
	}
}

func testIndexCanStopStartIndex(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	indexDef := index.createIndexDefinition()
	{
		op := NewPutIndexesOperation(indexDef)
		err = store.maintenance().send(op)
		assert.NoError(t, err)
	}

	{
		op := NewStopIndexOperation(indexDef.getName())
		err = store.maintenance().send(op)
		assert.NoError(t, err)
	}

	{
		op := NewGetIndexingStatusOperation()
		err = store.maintenance().send(op)
		assert.NoError(t, err)
		indexingStatus := op.Command.Result
		assert.Equal(t, indexingStatus.getStatus(), IndexRunningStatus_RUNNING)
		indexStatus := indexingStatus.getIndexes()[0]
		assert.Equal(t, indexStatus.getStatus(), IndexRunningStatus_PAUSED)
	}

	{
		op := NewStartIndexOperation(indexDef.getName())
		err = store.maintenance().send(op)
		assert.NoError(t, err)
	}

	{
		op := NewGetIndexingStatusOperation()
		err = store.maintenance().send(op)
		assert.NoError(t, err)
		indexingStatus := op.Command.Result
		assert.Equal(t, indexingStatus.getStatus(), IndexRunningStatus_RUNNING)
		indexStatus := indexingStatus.getIndexes()[0]
		assert.Equal(t, indexStatus.getStatus(), IndexRunningStatus_RUNNING)
	}
}

func testIndexCanSetIndexLockMode(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	indexDef := index.createIndexDefinition()
	{
		op := NewPutIndexesOperation(indexDef)
		err = store.maintenance().send(op)
		assert.NoError(t, err)
	}

	{
		op := NewSetIndexesLockOperation(indexDef.getName(), IndexLockMode_LOCKED_ERROR)
		err = store.maintenance().send(op)
		assert.NoError(t, err)
	}

	{
		op := NewGetIndexOperation(indexDef.getName())
		err = store.maintenance().send(op)
		newIndexDef := op.Command.Result
		assert.Equal(t, *newIndexDef.getLockMode(), IndexLockMode_LOCKED_ERROR)
	}
}

func testIndexCanSetIndexPriority(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	indexDef := index.createIndexDefinition()
	op := NewPutIndexesOperation(indexDef)
	err = store.maintenance().send(op)
	assert.NoError(t, err)

	op2 := NewSetIndexesPriorityOperation(indexDef.getName(), IndexPriority_HIGH)
	err = store.maintenance().send(op2)
	assert.NoError(t, err)

	op3 := NewGetIndexOperation(indexDef.getName())
	err = store.maintenance().send(op3)
	newIndexDef := op3.Command.Result
	assert.Equal(t, *newIndexDef.getPriority(), IndexPriority_HIGH)
}

func testIndexCanListErrors(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersInvalidIndex()
	indexDef := index.createIndexDefinition()
	op := NewPutIndexesOperation(indexDef)
	err = store.maintenance().send(op)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		user := NewUser()
		//user.setName(null)
		user.setAge(0)
		err = session.StoreEntity(user)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	err = gRavenTestDriver.waitForIndexing(store, store.getDatabase(), 0)
	assert.NoError(t, err)

	op2 := NewGetIndexErrorsOperation(nil)
	err = store.maintenance().send(op2)
	assert.NoError(t, err)
	indexErrors := op2.Command.Result
	assert.Equal(t, len(indexErrors), 1)

	op3 := NewGetIndexErrorsOperation([]string{indexDef.getName()})
	err = store.maintenance().send(op3)
	assert.NoError(t, err)
	perIndexErrors := op3.Command.Result
	assert.Equal(t, len(perIndexErrors), 1)
}

func testIndexCanGetIndexStatistics(t *testing.T) {

	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	userIndex := NewUsers_Index()
	err = userIndex.execute(store)
	assert.NoError(t, err)

	op := NewGetIndexesStatisticsOperation()
	err = store.maintenance().send(op)
	assert.NoError(t, err)

	indexStats := op.Command.Result
	assert.Equal(t, len(indexStats), 1)
}

func TestIndexOperations(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Recovered in %s\n", t.Name())
		}
		destroyDriver()
	}()

	// order matches Java tests
	testIndexHasIndexChanged(t)
	testIndexCanListErrors(t)
	testIndexCanGetIndexStatistics(t)
	testIndexCanSetIndexPriority(t)
	testIndexCanDisableAndEnableIndex(t)
	testIndexGetCanIndexes(t)
	testIndexCanDeleteIndex(t)
	testIndexCanStopStartIndexing(t)
	testIndexCanGetIndexStatistics(t)
	testIndexCanStopStartIndex(t)
	testIndexCanSetIndexLockMode(t)
	testIndexGetTerms(t)
}
