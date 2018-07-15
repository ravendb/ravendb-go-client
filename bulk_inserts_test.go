package ravendb

import (
	"testing"

	"github.com/ravendb/ravendb-go-client/pkg/proxy"
)

func bulkInsertsTest_simpleBulkInsertShouldWork(t *testing.T) {

}

func bulkInsertsTest_killedToEarly(t *testing.T) {

}

func bulkInsertsTest_shouldNotAcceptIdsEndingWithPipeLine(t *testing.T) {

}

func bulkInsertsTest_canModifyMetadataWithBulkInsert(t *testing.T) {

}

type FooBar struct {
	Name string
}

func (f *FooBar) getName() string {
	return f.Name
}

func (f *FooBar) setName(name string) {
	f.Name = name
}

func TestBulkInserts(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	if useProxy() {
		proxy.ChangeLogFile("trace_bulk_inserts_go.txt")
	}

	if false {
		dumpFailedHTTP = true
	}

	createTestDriver()
	defer deleteTestDriver()

	// matches order of Java tests
	bulkInsertsTest_simpleBulkInsertShouldWork(t)
	bulkInsertsTest_shouldNotAcceptIdsEndingWithPipeLine(t)
	bulkInsertsTest_killedToEarly(t)
	bulkInsertsTest_canModifyMetadataWithBulkInsert(t)
}
