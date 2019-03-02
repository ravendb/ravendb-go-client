package tests

import (
	"testing"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func NewUsersInvalidIndex() *ravendb.IndexCreationTask {
	res := ravendb.NewIndexCreationTask("UsersInvalidIndex")
	res.Map = "from u in docs.Users select new { a = 5 / u.Age }"
	return res
}

func NewUsers_Index() *ravendb.IndexCreationTask {
	res := ravendb.NewIndexCreationTask("Users_Index")
	res.Map = "from u in docs.Users select new { u.name }"
	return res
}

func testIndexCanDeleteIndex(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	err = index.Execute(store, nil, "")
	assert.NoError(t, err)

	op := ravendb.NewGetIndexNamesOperation(0, 10)
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)
	indexNames := op.Command.Result
	assert.True(t, stringArrayContains(indexNames, "UsersIndex"))

	op2 := ravendb.NewDeleteIndexOperation("UsersIndex")
	err = store.Maintenance().Send(op2)
	assert.NoError(t, err)

	op3 := ravendb.NewGetIndexNamesOperation(0, 10)
	err = store.Maintenance().Send(op3)
	assert.NoError(t, err)
	indexNames = op3.Command.Result
	assert.Equal(t, len(indexNames), 0)
}

func testIndexCanDisableAndEnableIndex(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	err = index.Execute(store, nil, "")
	assert.NoError(t, err)

	{
		op, err := ravendb.NewDisableIndexOperation("UsersIndex")
		assert.NoError(t, err)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	{
		op := ravendb.NewGetIndexingStatusOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		indexingStatus := op.Command.Result
		indexStatus := indexingStatus.Indexes[0]
		assert.Equal(t, indexStatus.Status, ravendb.IndexRunningStatusDisabled)
	}

	{
		op, err := ravendb.NewEnableIndexOperation("UsersIndex")
		assert.NoError(t, err)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	{
		op := ravendb.NewGetIndexingStatusOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		indexingStatus := op.Command.Result
		indexStatus := indexingStatus.Indexes[0]
		assert.Equal(t, indexStatus.Status, ravendb.IndexRunningStatusRunning)
	}
}

func testIndexGetCanIndexes(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	err = index.Execute(store, nil, "")
	assert.NoError(t, err)

	{
		op := ravendb.NewGetIndexesOperation(0, 10)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		indexDefinitions := op.Command.Result
		assert.Equal(t, len(indexDefinitions), 1)
	}
}

func testIndexGetCanIndexesStats(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	index := NewUsersIndex()
	err = index.Execute(store, nil, "")
	assert.NoError(t, err)

	{
		op := ravendb.NewGetIndexesStatisticsOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		indexStats := op.Command.Result
		assert.Equal(t, len(indexStats), 1)
	}
}

func testIndexGetTerms(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	err = index.Execute(store, nil, "")
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		user := &User{}
		user.setName("Marcin")
		err = session.Store(user)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	err = driver.waitForIndexing(store, store.GetDatabase(), 0)
	assert.NoError(t, err)

	{
		op, err := ravendb.NewGetTermsOperation("UsersIndex", "name", "", 0)
		assert.NoError(t, err)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		terms := op.Command.Result
		assert.Equal(t, len(terms), 1)
		assert.Equal(t, terms[0], "marcin")
	}
}

func testIndexHasIndexChanged(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	indexDef := index.CreateIndexDefinition()
	op := ravendb.NewPutIndexesOperation(indexDef)
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)
	op2 := ravendb.NewIndexHasChangedOperation(indexDef)
	err = store.Maintenance().Send(op2)
	assert.NoError(t, err)
	{
		cmd := op2.Command
		assert.False(t, cmd.Result)
	}
	indexDef.Maps = []string{"from users"}

	op3 := ravendb.NewIndexHasChangedOperation(indexDef)
	err = store.Maintenance().Send(op3)
	assert.NoError(t, err)
	{
		cmd := op3.Command
		assert.True(t, cmd.Result)
	}
}

func testIndexCanStopStartIndexing(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	indexDef := index.CreateIndexDefinition()
	{
		op := ravendb.NewPutIndexesOperation(indexDef)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	{
		op := ravendb.NewStopIndexingOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	{
		op := ravendb.NewGetIndexingStatusOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		indexingStatus := op.Command.Result
		indexStatus := indexingStatus.Indexes[0]
		assert.Equal(t, indexStatus.Status, ravendb.IndexRunningStatusPaused)
	}

	{
		op := ravendb.NewStartIndexingOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	{
		op := ravendb.NewGetIndexingStatusOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		indexingStatus := op.Command.Result
		indexStatus := indexingStatus.Indexes[0]
		assert.Equal(t, indexStatus.Status, ravendb.IndexRunningStatusRunning)
	}
}

func testIndexCanStopStartIndex(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	indexDef := index.CreateIndexDefinition()
	{
		op := ravendb.NewPutIndexesOperation(indexDef)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	{
		op, err := ravendb.NewStopIndexOperation(indexDef.Name)
		assert.NoError(t, err)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	{
		op := ravendb.NewGetIndexingStatusOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		indexingStatus := op.Command.Result
		assert.Equal(t, indexingStatus.Status, ravendb.IndexRunningStatusRunning)
		indexStatus := indexingStatus.Indexes[0]
		assert.Equal(t, indexStatus.Status, ravendb.IndexRunningStatusPaused)
	}

	{
		op, err := ravendb.NewStartIndexOperation(indexDef.Name)
		assert.NoError(t, err)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	{
		op := ravendb.NewGetIndexingStatusOperation()
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		indexingStatus := op.Command.Result
		assert.Equal(t, indexingStatus.Status, ravendb.IndexRunningStatusRunning)
		indexStatus := indexingStatus.Indexes[0]
		assert.Equal(t, indexStatus.Status, ravendb.IndexRunningStatusRunning)
	}
}

func testIndexCanSetIndexLockMode(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	indexDef := index.CreateIndexDefinition()
	{
		op := ravendb.NewPutIndexesOperation(indexDef)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	{
		op, err := ravendb.NewSetIndexesLockOperation(indexDef.Name, ravendb.IndexLockModeLockedError)
		assert.NoError(t, err)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
	}

	{
		op := ravendb.NewGetIndexOperation(indexDef.Name)
		err = store.Maintenance().Send(op)
		assert.NoError(t, err)
		newIndexDef := op.Command.Result
		assert.Equal(t, newIndexDef.LockMode, ravendb.IndexLockModeLockedError)
	}
}

func testIndexCanSetIndexPriority(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersIndex()
	indexDef := index.CreateIndexDefinition()
	op := ravendb.NewPutIndexesOperation(indexDef)
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)

	op2, err := ravendb.NewSetIndexesPriorityOperation(indexDef.Name, ravendb.IndexPriorityHigh)
	assert.NoError(t, err)
	err = store.Maintenance().Send(op2)
	assert.NoError(t, err)

	op3 := ravendb.NewGetIndexOperation(indexDef.Name)
	err = store.Maintenance().Send(op3)
	assert.NoError(t, err)
	newIndexDef := op3.Command.Result
	assert.Equal(t, newIndexDef.Priority, ravendb.IndexPriorityHigh)
}

func testIndexCanListErrors(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	index := NewUsersInvalidIndex()
	indexDef := index.CreateIndexDefinition()
	op := ravendb.NewPutIndexesOperation(indexDef)
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store)
		user := &User{}
		//user.setName(null)
		user.Age = 0
		err = session.Store(user)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}

	err = driver.waitForIndexing(store, store.GetDatabase(), 0)
	assert.NoError(t, err)

	op2 := ravendb.NewGetIndexErrorsOperation(nil)
	err = store.Maintenance().Send(op2)
	assert.NoError(t, err)
	indexErrors := op2.Command.Result
	assert.Equal(t, len(indexErrors), 1)

	op3 := ravendb.NewGetIndexErrorsOperation([]string{indexDef.Name})
	err = store.Maintenance().Send(op3)
	assert.NoError(t, err)
	perIndexErrors := op3.Command.Result
	assert.Equal(t, len(perIndexErrors), 1)
}

func testIndexCanGetIndexStatistics(t *testing.T, driver *RavenTestDriver) {

	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	userIndex := NewUsers_Index()
	err = userIndex.Execute(store, nil, "")
	assert.NoError(t, err)

	op := ravendb.NewGetIndexesStatisticsOperation()
	err = store.Maintenance().Send(op)
	assert.NoError(t, err)

	indexStats := op.Command.Result
	assert.Equal(t, len(indexStats), 1)
}

func TestIndexOperations(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// order matches Java tests
	testIndexHasIndexChanged(t, driver)
	testIndexCanListErrors(t, driver)
	testIndexCanGetIndexStatistics(t, driver)
	testIndexCanSetIndexPriority(t, driver)
	testIndexCanDisableAndEnableIndex(t, driver)
	testIndexGetCanIndexes(t, driver)
	testIndexCanDeleteIndex(t, driver)
	testIndexCanStopStartIndexing(t, driver)
	testIndexCanGetIndexStatistics(t, driver)
	testIndexCanStopStartIndex(t, driver)
	testIndexCanSetIndexLockMode(t, driver)
	testIndexGetTerms(t, driver)
}
