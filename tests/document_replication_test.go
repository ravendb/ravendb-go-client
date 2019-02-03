package tests

import (
	"fmt"
	"os"
	"testing"
	"time"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
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

func documentReplication_getConflictsResult_command_should_work_properly(t *testing.T, driver *RavenTestDriver) {
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

	{
		session := openSessionMust(t, source)

		user1 := &User{}
		user1.setName("Value")

		err = session.StoreWithID(user1, "docs/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, destination)

		user1 := &User{}
		user1.setName("Value2")

		err = session.StoreWithID(user1, "docs/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	driver.setupReplication(source, destination)

	{
		session := openSessionMust(t, source)

		user1 := &User{}
		user1.setName("marker")

		err = session.StoreWithID(user1, "marker")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	var user *User
	err = driver.waitForDocumentToReplicate(destination, &user, "marker", time.Millisecond*2090)
	assert.NoError(t, err)

	command := ravendb.NewGetConflictsCommand("docs/1")
	err = destination.GetRequestExecutor("").ExecuteCommand(command, nil)
	assert.NoError(t, err)
	results := command.Result.Results
	assert.Equal(t, len(results), 2)
	assert.NotEqual(t, results[0].ChangeVector, results[1].ChangeVector)
}

func documentReplication_shouldCreateConflictThenResolveIt(t *testing.T, driver *RavenTestDriver) {
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

	{
		session := openSessionMust(t, source)

		user1 := &User{}
		user1.setName("Value")

		err = session.StoreWithID(user1, "docs/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, destination)

		user1 := &User{}
		user1.setName("Value2")

		err = session.StoreWithID(user1, "docs/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	driver.setupReplication(source, destination)

	{
		session := openSessionMust(t, source)

		user1 := &User{}
		user1.setName("marker")

		err = session.StoreWithID(user1, "marker")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	var user *User
	err = driver.waitForDocumentToReplicate(destination, &user, "marker", time.Millisecond*2090)
	assert.NoError(t, err)

	command := ravendb.NewGetConflictsCommand("docs/1")
	err = destination.GetRequestExecutor("").ExecuteCommand(command, nil)
	assert.NoError(t, err)
	results := command.Result.Results
	assert.Equal(t, len(results), 2)
	assert.NotEqual(t, results[0].ChangeVector, results[1].ChangeVector)

	{
		session := openSessionMust(t, destination)

		var user1 *User
		err = session.Load(&user1, "docs/1")
		assert.Error(t, err)
		_, ok := err.(*ravendb.DocumentConflictError)
		if !ok {
			fmt.Printf("error is '%s' of type %T\n", err, err)
		}
		assert.True(t, ok)

		session.Close()
	}

	//now actually resolve the conflict
	//(resolve by using first variant)
	putCommand := ravendb.NewPutDocumentCommand("docs/1", nil, results[0].Doc)
	err = destination.GetRequestExecutor("").ExecuteCommand(putCommand, nil)
	assert.NoError(t, err)

	{
		session := openSessionMust(t, destination)

		var loadedUser *User
		err = session.Load(&loadedUser, "docs/1")
		assert.NoError(t, err)
		name, ok := jsonGetAsText(results[0].Doc, "name")
		assert.True(t, ok)
		assert.Equal(t, *loadedUser.Name, name)

		session.Close()
	}
}

func enableReplicationTests() bool {
	if os.Getenv("RAVEN_License") != "" {
		return true
	}
	if os.Getenv("RAVEN_License_Path") != "" {
		return true
	}
	return false
}

func TestDocumentReplication(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	if !enableReplicationTests() {
		fmt.Printf("Skipping TestDocumentReplication because RAVEN_License env variable is not set\n")
		return
	}

	// TODO: ensure order matches Java's order
	documentReplication_canReplicateDocument(t, driver)
	documentReplication_getConflictsResult_command_should_work_properly(t, driver)
	documentReplication_shouldCreateConflictThenResolveIt(t, driver)
}
