package tests

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type User5 struct {
	ID        string
	Name      string
	PartnerID string
	Email     string
	Tags      []string
	Age       int
	Active    bool
}

func cofi_can_cache_document_with_includes(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user := User5{
			Name: "Ayende",
		}
		err = session.Store(&user)
		assert.NoError(t, err)

		partner := User5{
			PartnerID: "user5s/1-A",
		}
		err = session.Store(&partner)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)
		var u *User5
		err = session.Include("PartnerID").Load(&u, "user5s/2-A")
		assert.NoError(t, err)
		assert.NotNil(t, u)

		// TODO: why SaveChanges() ?
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var u *User5
		err = session.Include("PartnerID").Load(&u, "user5s/2-A")
		assert.NoError(t, err)
		assert.NotNil(t, u)

		cache := session.Advanced().GetRequestExecutor().Cache
		cacheSize := cache.GetNumberOfItems()
		assert.Equal(t, 1, cacheSize)

		session.Close()
	}
}

func cofi_can_avoid_using_server_for_load_with_include_if_everything_is_in_session_cacheAsync(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user := User5{
			Name: "Ayende",
		}
		err = session.Store(&user)
		assert.NoError(t, err)

		partner := User5{
			PartnerID: "user5s/1-A",
		}
		err = session.Store(&partner)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)
		var user *User5
		err = session.Load(&user, "user5s/2-A")
		assert.NoError(t, err)
		assert.NotNil(t, user)

		err = session.Load(&user, user.PartnerID)
		assert.NoError(t, err)
		old := session.Advanced().GetNumberOfRequests()
		assert.NotNil(t, user)

		var newUser *User5
		err = session.Include("PartnerID").Load(&newUser, "user5s/2-A")
		assert.NoError(t, err)
		assert.NotNil(t, newUser)

		new := session.Advanced().GetNumberOfRequests()
		assert.Equal(t, new, old)

		session.Close()
	}

}

func cofi_can_avoid_using_server_for_load_with_include_if_everything_is_in_session_cacheLazy(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user := User5{
			Name: "Ayende",
		}
		err = session.Store(&user)
		assert.NoError(t, err)

		partner := User5{
			PartnerID: "user5s/1-A",
		}
		err = session.Store(&partner)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		utp := reflect.TypeOf(&User5{})
		advanced := session.Advanced()
		advanced.Lazily().Load(utp, "user5s/2-A", nil)
		advanced.Lazily().Load(utp, "user5s/1-A", nil)

		_, err = advanced.Eagerly().ExecuteAllPendingLazyOperations()
		assert.NoError(t, err)

		old := advanced.GetNumberOfRequests()

		result1 := advanced.Lazily().Include("PartnerId").Load(utp, "user5s/2-A")
		v, err := result1.GetValue()
		assert.NoError(t, err)
		u := v.(*User5)
		assert.NotNil(t, u)

		new := advanced.GetNumberOfRequests()
		assert.Equal(t, new, old)

		session.Close()
	}
}

func cofi_can_avoid_using_server_for_load_with_include_if_everything_is_in_session_cache(t *testing.T) {
	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		user := User5{
			Name: "Ayende",
		}
		err = session.Store(&user)
		assert.NoError(t, err)

		partner := User5{
			PartnerID: "user5s/1-A",
		}
		err = session.Store(&partner)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var user *User5
		err = session.Load(&user, "user5s/2-A")
		assert.NoError(t, err)

		var partner *User5
		err = session.Load(&partner, user.PartnerID)
		assert.NoError(t, err)

		old := session.Advanced().GetNumberOfRequests()

		var res *User5
		err = session.Include("PartnerID").Load(&res, "user5s/2-A")
		assert.NoError(t, err)

		new := session.Advanced().GetNumberOfRequests()
		assert.Equal(t, old, new)

		session.Close()
	}
}

func cofi_can_avoid_using_server_for_multiload_with_include_if_everything_is_in_session_cache(t *testing.T) {
}

func TestCachingOfDocumentInclude(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	cofi_can_avoid_using_server_for_multiload_with_include_if_everything_is_in_session_cache(t)
	cofi_can_avoid_using_server_for_load_with_include_if_everything_is_in_session_cacheLazy(t)
	cofi_can_cache_document_with_includes(t)
	cofi_can_avoid_using_server_for_load_with_include_if_everything_is_in_session_cacheAsync(t)
	cofi_can_avoid_using_server_for_load_with_include_if_everything_is_in_session_cache(t)
}
