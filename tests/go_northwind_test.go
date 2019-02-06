package tests

import (
	"fmt"
	"github.com/ravendb/ravendb-go-client"
	"github.com/ravendb/ravendb-go-client/examples/northwind"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

// Note: those are ad-hoc tests as they require http://live-test.ravendb.net
// to be running. They are not run in CI but can be run locally

func getLiveTestStoreMust(t *testing.T) *ravendb.DocumentStore {
	serverNodes := []string{"http://live-test.ravendb.net"}
	store := ravendb.NewDocumentStore(serverNodes, "Demo")
	err := store.Initialize()
	must(err)
	return store
}

func goNorthwindEmployeeLoad(t *testing.T, driver *RavenTestDriver) {
	var err error

	store := getLiveTestStoreMust(t)
	defer store.Close()

	session, err := store.OpenSession("")
	assert.NoError(t, err)
	defer session.Close()

	var e *northwind.Employee
	err = session.Load(&e, "employees/7-A")
	assert.NoError(t, err)
	fmt.Printf("employee: %#v\n", e)
}

func TestGoNorthwind(t *testing.T) {
	// t.Parallel()
	// not enabled in CI, only when run from run_single_test.ps1 and similar
	if os.Getenv("ENABLE_NORTHWIND_TESTS") == "" {
		fmt.Printf("Skipping TestGoNorthwind because ENABLE_NORTHWIND_TESTS env is not set\n")
		return
	}

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	goNorthwindEmployeeLoad(t, driver)
}
