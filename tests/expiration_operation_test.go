package tests

import (
	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
	"testing"
)

func expirationConfigurationOperation(t *testing.T, driver *RavenTestDriver) {
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	configureExpiration := ravendb.ExpirationConfiguration{
		Disabled: false,
	}

	opExpiration, err := ravendb.NewConfigureExpirationOperationWithConfiguration(&configureExpiration)
	assert.NoError(t, err)

	err = store.Maintenance().Send(opExpiration)
	assert.NoError(t, err)

	lastRaftEtag := *opExpiration.Command.Result.RaftCommandIndex
	assert.NotNil(t, lastRaftEtag)

	var deleteFrequency int64 = 60
	opExpiration, err = ravendb.NewConfigureExpirationOperation(false, &deleteFrequency, nil)
	assert.NoError(t, err)

	err = store.Maintenance().Send(opExpiration)
	assert.NoError(t, err)

	assert.NotNil(t, *opExpiration.Command.Result.RaftCommandIndex)
	assert.NotEqual(t, lastRaftEtag, *opExpiration.Command.Result.RaftCommandIndex)
}

func TestExpirationConfiguration(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)
	expirationConfigurationOperation(t, driver)
}
