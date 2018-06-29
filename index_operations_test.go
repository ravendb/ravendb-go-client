package ravendb

import (
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

func testIndexCanDeleteIndex(t *testing.T) {
}

func testIndexCanDisableAndEnableIndex(t *testing.T) {
}

func testIndexGetCanIndexes(t *testing.T) {
}
func testIndexGetCanIndexesStats(t *testing.T) {
}
func testIndexGetTerms(t *testing.T) {
}
func testIndexHasIndexChanged(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
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
}
func testIndexCanStopStartIndex(t *testing.T) {
}
func testIndexCanSetIndexLockMode(t *testing.T) {
}
func testIndexCanSetIndexPriority(t *testing.T) {
}
func testIndexCanListErrors(t *testing.T) {
}
func testIndexCanGetIndexStatistics(t *testing.T) {
}

func TestIndexOperations(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	if useProxy() {
		proxy.ChangeLogFile("trace_index_operations_go.txt")
	}

	createTestDriver()
	defer deleteTestDriver()

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
