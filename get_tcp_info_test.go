package ravendb

import (
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
	"github.com/stretchr/testify/assert"
)

func canGetTcpInfo(t *testing.T) {
	store := getDocumentStoreMust(t)
	command := NewGetTcpInfoCommand("test")
	err := store.GetRequestExecutor().executeCommand(command)
	assert.NoError(t, err)
	result := command.getResult().(*TcpConnectionInfo)
	assert.NotNil(t, result)
	assert.Nil(t, result.getCertificate())
	// Note: in Java this tests for non-nil but Port is not sent
	// in Json, so don't quite understand that. Unless Java check
	// is bogus
	assert.Equal(t, 0, result.getPort())
	assert.NotEmpty(t, result.getUrl())
}

func TestGetTcpInfo(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	if useProxy() {
		proxy.ChangeLogFile("trace_get_tcp_info_go.txt")
	}
	canGetTcpInfo(t)
}
