package tests

import (
	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func documentReplication_canReplicateDocument(t *testing.T, driver *RavenTestDriver) {
	driver.customize = func(r *ravendb.DatabaseRecord) {
		conflictSolver := &ravendb.ConflictSolver{
			ResolveToLatest:     false,
			ResolveByCollection: map[string]*ravendb.ScriptResolver{},
		}
		r.ConflictSolverConfig = conflictSolver
	}
	defer func() {
		driver.customize = nil
	}()

	var err error
	source := driver.getDocumentStoreMust(t)
	defer source.Close()

	destination := driver.getDocumentStoreMust(t)
	defer destination.Close()

	driver.setupReplication(source, destination)
	var id string

	{
		session := openSessionMust(t, source)

		user := &User{}
		user.setName("Arek")

		err = session.Store(user)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		id = user.ID

		session.Close()
	}
	var fetchedUser *User
	err = driver.waitForDocumentToReplicate(destination, &fetchedUser, id, time.Millisecond*10000)
	assert.NoError(t, err)
	assert.Equal(t, *fetchedUser.Name, "Arek")
}

func TestDocumentReplication(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// TODO: ensure order matches Java's order
	//documentReplication_canReplicateDocument(t, driver)
}
