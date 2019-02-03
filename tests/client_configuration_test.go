package tests

import (
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func clientConfigurationCanHandleNoConfiguration(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	operation := ravendb.NewGetClientConfigurationOperation()
	err := store.Maintenance().Send(operation)
	assert.NoError(t, err)
	result := operation.Command.Result
	assert.Nil(t, result.Configuration)
	assert.True(t, result.Etag > 0)
}

func clientConfigurationCanSaveAndReadClientConfiguration(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	configurationToSave := &ravendb.ClientConfiguration{
		Etag:                          123,
		MaxNumberOfRequestsPerSession: 80,
		ReadBalanceBehavior:           ravendb.ReadBalanceBehaviorFastestNode,
		IsDisabled:                    true,
	}

	saveOperation, err := ravendb.NewPutClientConfigurationOperation(configurationToSave)
	assert.NoError(t, err)
	store.Maintenance().Send(saveOperation)
	operation := ravendb.NewGetClientConfigurationOperation()
	err = store.Maintenance().Send(operation)
	assert.NoError(t, err)
	result := operation.Command.Result
	assert.True(t, result.Etag > 0)
	newConfiguration := result.Configuration
	assert.NotNil(t, newConfiguration)
	assert.True(t, newConfiguration.Etag > configurationToSave.Etag)
	assert.True(t, newConfiguration.IsDisabled)
	assert.Equal(t, newConfiguration.MaxNumberOfRequestsPerSession, 80)
	assert.Equal(t, newConfiguration.ReadBalanceBehavior, ravendb.ReadBalanceBehaviorFastestNode)
}

func TestClientConfiguration(t *testing.T) {
	// // t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	clientConfigurationCanHandleNoConfiguration(t, driver)
	clientConfigurationCanSaveAndReadClientConfiguration(t, driver)
}
