package tests

import (
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func clientConfiguration_canHandleNoConfiguration(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	operation := ravendb.NewGetClientConfigurationOperation()
	err := store.Maintenance().Send(operation)
	assert.NoError(t, err)
	result := operation.Command.Result
	assert.Nil(t, result.GetConfiguration())
	//TODO: java checks that result.getEtag() is not nil, which does not apply
}

func clientConfiguration_canSaveAndReadClientConfiguration(t *testing.T, driver *RavenTestDriver) {
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	configurationToSave := &ravendb.ClientConfiguration{
		Etag:                          123,
		MaxNumberOfRequestsPerSession: 80,
		ReadBalanceBehavior:           ravendb.ReadBalanceBehavior_FASTEST_NODE,
		IsDisabled:                    true,
	}

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
	assert.True(t, newConfiguration.Etag > configurationToSave.Etag)
	assert.True(t, newConfiguration.IsDisabled)
	assert.Equal(t, newConfiguration.MaxNumberOfRequestsPerSession, 80)
	assert.Equal(t, newConfiguration.ReadBalanceBehavior, ravendb.ReadBalanceBehavior_FASTEST_NODE)
}

func TestClientConfiguration(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	clientConfiguration_canHandleNoConfiguration(t, driver)
	clientConfiguration_canSaveAndReadClientConfiguration(t, driver)
}
