package tests

import (
	"testing"

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
	}

}

func TestGo1(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	go1Test(t, driver)
}
