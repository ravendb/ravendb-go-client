package ravendb

import (
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
)

func NewUsersIndex() *AbstractIndexCreationTask {
	res := NewAbstractIndexCreationTask("UsersIndex")
	res.smap = "from user in docs.users select new { user.name }"
	return res
}
func TestIndexesFromClient(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	if useProxy() {
		proxy.ChangeLogFile("trace_indexes_from_client_go.txt")
	}

	createTestDriver()
	defer deleteTestDriver()

	// order matches Java tests
}
