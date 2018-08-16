package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

func clientConfiguration_canHandleNoConfiguration(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	operation := ravendb.NewGetClientConfigurationOperation()
	err := store.Maintenance().Send(operation)
	assert.NoError(t, err)
	result := operation.Command.Result
	assert.Nil(t, result.GetConfiguration())
	//TODO: java checks that result.getEtag() is not nil, which does not apply
}

func clientConfiguration_canSaveAndReadClientConfiguration(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	configurationToSave := ravendb.NewClientConfiguration()
	configurationToSave.SetEtag(123)
	configurationToSave.SetMaxNumberOfRequestsPerSession(80)
	configurationToSave.SetReadBalanceBehavior(ravendb.ReadBalanceBehavior_FASTEST_NODE)
	configurationToSave.SetDisabled(true)

	saveOperation, err := ravendb.NewPutClientConfigurationOperation(configurationToSave)
	assert.NoError(t, err)
	store.Maintenance().Send(saveOperation)
	operation := ravendb.NewGetClientConfigurationOperation()
	err = store.Maintenance().Send(operation)
	assert.NoError(t, err)
	result := operation.Command.Result
	assert.True(t, result.GetEtag() > 0)
	newConfiguration := result.GetConfiguration()
	assert.NotNil(t, newConfiguration)
	assert.True(t, newConfiguration.GetEtag() > configurationToSave.GetEtag())
	assert.True(t, newConfiguration.IsDisabled())
	assert.Equal(t, newConfiguration.GetMaxNumberOfRequestsPerSession(), 80)
	assert.Equal(t, newConfiguration.GetReadBalanceBehavior(), ravendb.ReadBalanceBehavior_FASTEST_NODE)
}

func TestClientConfiguration(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	clientConfiguration_canHandleNoConfiguration(t)
	clientConfiguration_canSaveAndReadClientConfiguration(t)
}
