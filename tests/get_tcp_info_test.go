package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

func getTcpInfoTest_canGetTcpInfo(t *testing.T) {
	store := getDocumentStoreMust(t)
	defer store.Close()

	command := ravendb.NewGetTcpInfoCommand("test")
	err := store.GetRequestExecutor().ExecuteCommand(command)
	assert.NoError(t, err)
	result := command.Result
	assert.NotNil(t, result)
	assert.Nil(t, result.GetCertificate())
	// Note: in Java this tests for non-nil but Port is not sent
	// in Json, so don't quite understand that. Unless Java check
	// is bogus
	assert.Equal(t, 0, result.GetPort())
	assert.NotEmpty(t, result.GetUrl())
}

func TestGetTcpInfo(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	getTcpInfoTest_canGetTcpInfo(t)
}
