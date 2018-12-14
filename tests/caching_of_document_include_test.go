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

func cofi_can_cache_document_with_includes(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
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

func cofi_can_avoid_using_server_for_load_with_include_if_everything_is_in_session_cacheAsync(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
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

func cofi_can_avoid_using_server_for_load_with_include_if_everything_is_in_session_cacheLazy(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
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

func cofi_can_avoid_using_server_for_load_with_include_if_everything_is_in_session_cache(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
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

func cofi_can_avoid_using_server_for_multiload_with_include_if_everything_is_in_session_cache(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		names := []string{"Additional", "Ayende", "Michael", "Fitzhak", "Maxim"}
		for _, name := range names {
			user := &User5{
				Name: name,
			}
			err = session.Store(user)
			assert.NoError(t, err)
		}

		withPartner := &User5{
			PartnerID: "user5s/1-A",
		}
		err = session.Store(withPartner)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var u2, u6 *User5
		err = session.Load(&u2, "user5s/2-A")
		assert.NoError(t, err)
		err = session.Load(&u6, "user5s/6-A")
		assert.NoError(t, err)

		inp := []string{"user5s/1-A", "user5s/2-A", "user5s/3-A", "user5s/4-A", "user5s/5-A", "user5s/6-A"}
		u4 := make(map[string]*User5)
		err = session.LoadMulti(u4, inp)
		assert.NoError(t, err)

		var u *User5
		err = session.Load(&u, u6.PartnerID)
		assert.NoError(t, err)

		old := session.Advanced().GetNumberOfRequests()

		res := make(map[string]*User5)
		ids := []string{"user5s/2-A", "user5s/3-A", "user5s/6-A"}
		err = session.Include("PartnerID").LoadMulti(res, ids)
		assert.NoError(t, err)

		new := session.Advanced().GetNumberOfRequests()
		assert.Equal(t, old, new)

		session.Close()
	}
}

func TestCachingOfDocumentInclude(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	//cofi_can_avoid_using_server_for_multiload_with_include_if_everything_is_in_session_cacheg(t)
	cofi_can_avoid_using_server_for_load_with_include_if_everything_is_in_session_cacheLazy(t, driver)
	cofi_can_cache_document_with_includes(t, driver)
	cofi_can_avoid_using_server_for_load_with_include_if_everything_is_in_session_cacheAsync(t, driver)
	cofi_can_avoid_using_server_for_load_with_include_if_everything_is_in_session_cache(t, driver)
}
