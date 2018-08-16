package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

func getStatisticsCommandTest_canGetStats(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	executor := store.GetRequestExecutor()

	sampleData := NewCreateSampleDataOperation()
	err = store.Maintenance().Send(sampleData)
	assert.NoError(t, err)

	err = gRavenTestDriver.waitForIndexing(store, store.GetDatabase(), 0)
	assert.NoError(t, err)
	command := ravendb.NewGetStatisticsCommand()
	err = executor.ExecuteCommand(command)
	assert.NoError(t, err)
	stats := command.Result
	assert.NotNil(t, stats)
	assert.True(t, stats.GetLastDocEtag() > 0)
	assert.Equal(t, stats.GetCountOfIndexes(), 3)
	assert.Equal(t, stats.GetCountOfDocuments(), 1059)
	assert.True(t, stats.GetCountOfRevisionDocuments() > 0)
	assert.Equal(t, stats.GetCountOfDocumentsConflicts(), 0)
	assert.Equal(t, stats.GetCountOfConflicts(), 0)
	assert.Equal(t, stats.GetCountOfUniqueAttachments(), 17)
	assert.NotEqual(t, stats.GetDatabaseChangeVector(), "")
	assert.NotEqual(t, stats.GetDatabaseID(), "")
	assert.NotNil(t, stats.GetPager())
	assert.NotNil(t, stats.GetLastIndexingTime())
	assert.NotNil(t, stats.GetIndexes())
	assert.NotEqual(t, stats.GetSizeOnDisk().GetHumaneSize(), "")
	assert.NotEqual(t, stats.GetSizeOnDisk().GetSizeInBytes(), 0)

	indexes := stats.GetIndexes()
	for _, indexInformation := range indexes {
		assert.NotEqual(t, indexInformation.GetName(), "")
		assert.False(t, indexInformation.IsStale())
		assert.NotNil(t, indexInformation.GetState())
		assert.NotEqual(t, indexInformation.GetLockMode(), "")
		assert.NotEqual(t, indexInformation.GetPriority(), "")
		assert.NotEqual(t, indexInformation.GetType(), "")
		assert.NotNil(t, indexInformation.GetLastIndexingTime())
	}
}

func TestGetStatisticsCommand(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	getStatisticsCommandTest_canGetStats(t)
}
