package tests

import (
	"testing"
	"time"

	ravendb "github.com/ravendb/ravendb-go-client"

	"github.com/stretchr/testify/assert"
)

func assertIllegalArgumentError2(t *testing.T, err error, exp string) {
	assert.Error(t, err)
	if err != nil {
		_, ok := err.(*ravendb.IllegalArgumentError)
		assert.True(t, ok)
		assert.Equal(t, exp, err.Error())
	}
}

func assertIllegalArgumentError(t *testing.T, err error) {
	assert.Error(t, err)
	if err != nil {
		_, ok := err.(*ravendb.IllegalArgumentError)
		assert.True(t, ok)
	}
}

func go1Test(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	session := openSessionMust(t, store)
	user := User{}

	// check validation of arguments to Store and Delete

	{
		// can't store/delete etc. nil
		var v interface{}
		err = session.Store(v)
		assertIllegalArgumentError2(t, err, "entity can't be nil")
		err = session.StoreWithID(v, "users/1")
		assertIllegalArgumentError(t, err)
		err = session.DeleteEntity(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetMetadataFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetChangeVectorFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetLastModifiedFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.HasChanged(v)
		assertIllegalArgumentError(t, err)
		err = session.Evict(v)
		assertIllegalArgumentError(t, err)
	}

	{
		// can't store/delete etc. nil pointer
		var v *User
		err = session.Store(v)
		assertIllegalArgumentError2(t, err, "entity of type *tests.User can't be nil")
		err = session.StoreWithID(v, "users/1")
		assertIllegalArgumentError(t, err)
		err = session.DeleteEntity(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetMetadataFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetChangeVectorFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetLastModifiedFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.HasChanged(v)
		assertIllegalArgumentError(t, err)
		err = session.Evict(v)
		assertIllegalArgumentError(t, err)
	}

	{
		// can't store/delete etc. struct
		v := user
		err = session.Store(v)
		assertIllegalArgumentError2(t, err, "entity can't be of type User, try passing *User")
		err = session.StoreWithID(v, "users/1")
		assertIllegalArgumentError(t, err)
		err = session.DeleteEntity(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetMetadataFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetChangeVectorFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetLastModifiedFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.HasChanged(v)
		assertIllegalArgumentError(t, err)
		err = session.Evict(v)
		assertIllegalArgumentError(t, err)
	}

	{
		// can't store/delete etc. **struct (double pointer values)
		ptrUser := &user
		v := &ptrUser
		err = session.Store(v)
		assertIllegalArgumentError2(t, err, "entity can't be of type **tests.User, try passing *tests.User")
		err = session.StoreWithID(v, "users/1")
		assertIllegalArgumentError(t, err)
		err = session.DeleteEntity(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetMetadataFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetChangeVectorFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetLastModifiedFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.HasChanged(v)
		assertIllegalArgumentError(t, err)
		err = session.Evict(v)
		assertIllegalArgumentError(t, err)
	}

	{
		// can't store/delete etc. a nil map
		var v map[string]interface{}
		err = session.Store(v)
		assertIllegalArgumentError2(t, err, "entity can't be a nil map")
		err = session.StoreWithID(v, "users/1")
		assertIllegalArgumentError(t, err)
		err = session.DeleteEntity(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetMetadataFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetChangeVectorFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetLastModifiedFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.HasChanged(v)
		assertIllegalArgumentError(t, err)
		err = session.Evict(v)
		assertIllegalArgumentError(t, err)
	}

	{
		// can't store/delete etc. *map[string]interface{}
		m := map[string]interface{}{}
		v := &m
		err = session.Store(v)
		assertIllegalArgumentError2(t, err, "entity can't be of type *map[string]interface {}, try passing map[string]interface {}")
		err = session.StoreWithID(v, "users/1")
		assertIllegalArgumentError(t, err)
		err = session.DeleteEntity(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetMetadataFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetChangeVectorFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetLastModifiedFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.HasChanged(v)
		assertIllegalArgumentError(t, err)
		err = session.Evict(v)
		assertIllegalArgumentError(t, err)
	}

}

func goStore(t *testing.T, session *ravendb.DocumentSession) []*User {
	var err error
	var res []*User
	{
		names := []string{"John", "Mary", "Paul"}
		for _, name := range names {
			u := &User{}
			u.setName(name)
			err := session.Store(u)
			assert.NoError(t, err)
			res = append(res, u)
		}
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}
	return res
}

func goTestGetLastModifiedForAndChanges(t *testing.T, driver *RavenTestDriver) {
	var err error
	var changed, hasChanges bool

	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	var users []*User
	var lastModifiedFirst *time.Time
	{
		session := openSessionMust(t, store)
		users = goStore(t, session)
		lastModifiedFirst, err = session.GetLastModifiedFor(users[0])
		assert.NoError(t, err)
		assert.NotNil(t, lastModifiedFirst)
		session.Close()
	}

	{
		session := openSessionMust(t, store)

		// test HasChanges()
		hasChanges = session.HasChanges()
		assert.False(t, hasChanges)

		var u *User
		id := users[0].ID
		err = session.Load(&u, id)
		assert.NoError(t, err)
		assert.Equal(t, id, u.ID)
		lastModified, err := session.GetLastModifiedFor(u)
		assert.NoError(t, err)
		assert.Equal(t, *lastModifiedFirst, *lastModified)

		changed, err = session.HasChanged(u)
		assert.NoError(t, err)
		assert.False(t, changed)

		// check last modified changes after modification
		u.Age = 5
		err = session.Store(u)
		assert.NoError(t, err)

		changed, err = session.HasChanged(u)
		assert.NoError(t, err)
		assert.True(t, changed)

		hasChanges = session.HasChanges()
		assert.True(t, hasChanges)

		err = session.SaveChanges()
		assert.NoError(t, err)

		lastModified, err = session.GetLastModifiedFor(u)
		assert.NoError(t, err)
		diff := (*lastModified).Sub(*lastModifiedFirst)
		assert.True(t, diff > 0)

		session.Close()
	}

	{

		// test HasChanged() detects deletion
		session := openSessionMust(t, store)
		var u *User
		id := users[0].ID
		err = session.Load(&u, id)
		assert.NoError(t, err)

		err = session.DeleteEntity(u)
		assert.NoError(t, err)

		/*
			// TODO: should deleted items be reported as changed?
			changed, err = session.HasChanged(u)
			assert.NoError(t, err)
			assert.True(t, changed)
		*/

		hasChanges = session.HasChanges()
		assert.True(t, hasChanges)

		// Evict undoes deletion so we shouldn't have changes
		err = session.Evict(u)
		assert.NoError(t, err)

		changed, err = session.HasChanged(u)
		assert.NoError(t, err)
		assert.False(t, changed)

		hasChanges = session.HasChanges()
		assert.False(t, hasChanges)
	}
}

func TestGo1(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	go1Test(t, driver)
	goTestGetLastModifiedForAndChanges(t, driver)
}
