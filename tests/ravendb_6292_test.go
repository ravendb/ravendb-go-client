package tests

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	ravendb "github.com/ravendb/ravendb-go-client"
)

func ravendb6292_ifIncludedDocumentIsConflictedItShouldNotThrowConflictException(t *testing.T, driver *RavenTestDriver) {
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
	store1 := driver.getDocumentStoreMust(t)
	defer store1.Close()

	store2 := driver.getDocumentStoreMust(t)
	defer store2.Close()

	{
		session := openSessionMust(t, store1)

		address := &Address{
			City: "New York",
		}
		err = session.StoreWithID(address, "addresses/1")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store2)
		address := &Address{
			City: "Torun",
		}

		err = session.StoreWithID(address, "addresses/1")
		assert.NoError(t, err)

		user := &User{}
		user.setName("John")
		user.AddressID = "addresses/1"
		err = session.StoreWithID(user, "users/1")
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	driver.setupReplication(store1, store2)

	var addressTmp *Address
	err = driver.waitForConflict(store2, &addressTmp, "addresses/1")
	if err != nil {
		fmt.Printf("Got unexpected error '%s' of type %T\n", err, err)
	}
	assert.NoError(t, err)

	{
		session := openSessionMust(t, store2)

		clazz := reflect.TypeOf(&User{})
		documentQuery := session.Advanced().DocumentQueryOld(clazz).Include("addressId")

		iq := documentQuery.GetIndexQuery()

		var user *User
		err = documentQuery.First(&user)
		assert.NoError(t, err)
		assert.Equal(t, *user.Name, "John")

		{
			var addr *Address
			err = session.Load(&addr, user.AddressID)
			assert.Error(t, err)
			_, ok := err.(*ravendb.DocumentConflictError)
			assert.True(t, ok)
		}

		queryCommand, err := ravendb.NewQueryCommand(ravendb.NewDocumentConventions(), iq, false, false)
		assert.NoError(t, err)

		err = store2.GetRequestExecutor("").ExecuteCommand(queryCommand)
		assert.NoError(t, err)

		result := queryCommand.Result
		address := result.Includes["addresses/1"].(map[string]interface{})

		metadata := address["@metadata"].(map[string]interface{})
		id := metadata["@id"].(string)
		assert.Equal(t, id, "addresses/1")

		assert.True(t, ravendb.JSONExtensionsTryGetConflict(metadata))
	}
}

func (d *RavenTestDriver) waitForConflict(store *ravendb.DocumentStore, result interface{}, id string) error {
	sw := time.Now()
	timeout := time.Millisecond * 10090
	for {
		time.Sleep(time.Millisecond * 500)
		dur := time.Since(sw)
		if dur > timeout {
			return nil
		}
		{
			session, err := store.OpenSession()
			if err != nil {
				return err
			}
			err = session.Load(result, id)
			session.Close()
			if err != nil {
				// Note: in Java the code checks for ConflictException which is a base class
				// for ConcurrencyException and DocumentConflictException
				if _, ok := err.(*ravendb.ConflictError); ok {
					return nil
				}
				if _, ok := err.(*ravendb.ConcurrencyError); ok {
					return nil
				}
				if _, ok := err.(*ravendb.DocumentConflictError); ok {
					return nil
				}
				return err
			}
		}
	}
	return fmt.Errorf("Waited for conflict on '%s' but it did not happen", id)
}

func TestRavenDB6292(t *testing.T) {
	// t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	if !enableReplicationTests() {
		fmt.Printf("Skipping TestDocumentReplication because RAVEN_License env variable is not set\n")
		return
	}

	// matches Java's order
	ravendb6292_ifIncludedDocumentIsConflictedItShouldNotThrowConflictException(t, driver)
}
