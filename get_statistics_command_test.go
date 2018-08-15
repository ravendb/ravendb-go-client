package ravendb

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
	command := NewGetStatisticsCommand()
	err = executor.ExecuteCommand(command)
	assert.NoError(t, err)
	stats := command.Result
	assert.NotNil(t, stats)
	assert.True(t, stats.getLastDocEtag() > 0)
	assert.Equal(t, stats.getCountOfIndexes(), 3)
	assert.Equal(t, stats.getCountOfDocuments(), 1059)
	assert.True(t, stats.getCountOfRevisionDocuments() > 0)
	assert.Equal(t, stats.getCountOfDocumentsConflicts(), 0)
	assert.Equal(t, stats.getCountOfConflicts(), 0)
	assert.Equal(t, stats.getCountOfUniqueAttachments(), 17)
	assert.NotEqual(t, stats.getDatabaseChangeVector(), "")
	assert.NotEqual(t, stats.getDatabaseId(), "")
	assert.NotNil(t, stats.getPager())
	assert.NotNil(t, stats.getLastIndexingTime())
	assert.NotNil(t, stats.getIndexes())
	assert.NotEqual(t, stats.getSizeOnDisk().getHumaneSize(), "")
	assert.NotEqual(t, stats.getSizeOnDisk().getSizeInBytes(), 0)

	indexes := stats.getIndexes()
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
