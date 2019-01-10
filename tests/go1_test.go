package tests

import (
	"testing"
	"time"

	"github.com/ravendb/ravendb-go-client"

	"github.com/stretchr/testify/assert"
)

func assertIllegalArgumentError(t *testing.T, err error, s ...string) {
	assert.Error(t, err)
	if err != nil {
		_, ok := err.(*ravendb.IllegalArgumentError)
		if !ok {
			assert.True(t, ok, "expected error of type *ravendb.IllegalArgumentError, got %T", err)
			return
		}
		if len(s) > 0 {
			panicIf(len(s) > 1, "only 0 or 1 strings are expected as s")
			assert.Equal(t, s[0], err.Error())
		}
	}
}

func assertIllegalStateError(t *testing.T, err error, s ...string) {
	assert.Error(t, err)
	if err != nil {
		_, ok := err.(*ravendb.IllegalStateError)
		if !ok {
			assert.True(t, ok, "expected error of type *ravendb.IllegalStateError, got %T", err)
			return
		}
		if len(s) > 0 {
			panicIf(len(s) > 1, "only 0 or 1 strings are expected as s")
			assert.Equal(t, s[0], err.Error())
		}
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
		assertIllegalArgumentError(t, err, "entity can't be nil")
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
		err = session.Advanced().PatchEntity(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().IncrementEntity(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().PatchArrayInEntity(v, "foo", nil)
		assertIllegalArgumentError(t, err)
		err = session.Refresh(v)
		assertIllegalArgumentError(t, err)
	}

	{
		// can't store/delete etc. nil pointer
		var v *User
		err = session.Store(v)
		assertIllegalArgumentError(t, err, "entity of type *tests.User can't be nil")
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
		err = session.Advanced().PatchEntity(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().IncrementEntity(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().PatchArrayInEntity(v, "foo", nil)
		assertIllegalArgumentError(t, err)
		err = session.Refresh(v)
		assertIllegalArgumentError(t, err)
	}

	{
		// can't store/delete etc. struct
		v := user
		err = session.Store(v)
		assertIllegalArgumentError(t, err, "entity can't be of type User, try passing *User")
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
		err = session.Advanced().PatchEntity(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().IncrementEntity(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().PatchArrayInEntity(v, "foo", nil)
		assertIllegalArgumentError(t, err)
		err = session.Refresh(v)
		assertIllegalArgumentError(t, err)
	}

	{
		// can't store/delete etc. **struct (double pointer values)
		ptrUser := &user
		v := &ptrUser
		err = session.Store(v)
		assertIllegalArgumentError(t, err, "entity can't be of type **tests.User, try passing *tests.User")
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
		err = session.Advanced().PatchEntity(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().IncrementEntity(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().PatchArrayInEntity(v, "foo", nil)
		assertIllegalArgumentError(t, err)
		err = session.Refresh(v)
		assertIllegalArgumentError(t, err)
	}

	{
		// can't store/delete etc. a nil map
		var v map[string]interface{}
		err = session.Store(v)
		assertIllegalArgumentError(t, err, "entity can't be a nil map")
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
		err = session.Advanced().PatchEntity(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().IncrementEntity(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().PatchArrayInEntity(v, "foo", nil)
		assertIllegalArgumentError(t, err)
		err = session.Refresh(v)
		assertIllegalArgumentError(t, err)
	}

	{
		// can't store/delete etc. *map[string]interface{}
		m := map[string]interface{}{}
		v := &m
		err = session.Store(v)
		assertIllegalArgumentError(t, err, "entity can't be of type *map[string]interface {}, try passing map[string]interface {}")
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
		err = session.Advanced().PatchEntity(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().IncrementEntity(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().PatchArrayInEntity(v, "foo", nil)
		assertIllegalArgumentError(t, err)
		err = session.Refresh(v)
		assertIllegalArgumentError(t, err)
	}

	{
		v := &User{} // dummy value that only has to pass type check
		adv := session.Advanced()

		err = adv.IncrementEntity(v, "", 1)
		assertIllegalArgumentError(t, err, "path can't be empty string")
		err = adv.IncrementEntity(v, "foo", nil)
		assertIllegalArgumentError(t, err, "valueToAdd can't be nil")

		err = adv.IncrementByID("", "foo", 1)
		assertIllegalArgumentError(t, err, "id can't be empty string")
		err = adv.IncrementByID("id", "", 1)
		assertIllegalArgumentError(t, err, "path can't be empty string")
		err = adv.IncrementByID("id", "foo", nil)
		assertIllegalArgumentError(t, err, "valueToAdd can't be nil")

		err = adv.PatchEntity(v, "", 1)
		assertIllegalArgumentError(t, err, "path can't be empty string")
		err = adv.PatchEntity(v, "foo", nil)
		assertIllegalArgumentError(t, err, "value can't be nil")

		err = adv.PatchByID("", "foo", 1)
		assertIllegalArgumentError(t, err, "id can't be empty string")
		err = adv.PatchByID("id", "", 1)
		assertIllegalArgumentError(t, err, "path can't be empty string")
		err = adv.PatchByID("id", "foo", nil)
		assertIllegalArgumentError(t, err, "value can't be nil")

		err = adv.PatchArrayInEntity(v, "", nil)
		assertIllegalArgumentError(t, err, "pathToArray can't be empty string")
		err = adv.PatchArrayInEntity(v, "foo", nil)
		assertIllegalArgumentError(t, err, "arrayAdder can't be nil")

		err = adv.PatchArrayByID("", "foo", nil)
		assertIllegalArgumentError(t, err, "id can't be empty string")
		err = adv.PatchArrayByID("id", "", nil)
		assertIllegalArgumentError(t, err, "pathToArray can't be empty string")
		err = adv.PatchArrayByID("id", "foo", nil)
		assertIllegalArgumentError(t, err, "arrayAdder can't be nil")
	}

	{
		_, err = session.Exists("")
		assertIllegalArgumentError(t, err, "id cannot be empty string")
	}

	session.Close()
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

func goTestListeners(t *testing.T, driver *RavenTestDriver) {
	var err error

	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	nBeforeStoreCalledCount := 0
	beforeStore := func(event *ravendb.BeforeStoreEventArgs) {
		_, ok := event.Entity.(*User)
		assert.True(t, ok)
		nBeforeStoreCalledCount++
	}
	beforeStoreID := store.AddBeforeStoreListener(beforeStore)

	nAfterSaveChangesCalledCount := 0
	afterSaveChanges := func(event *ravendb.AfterSaveChangesEventArgs) {
		_, ok := event.Entity.(*User)
		assert.True(t, ok)
		nAfterSaveChangesCalledCount++
	}
	afterSaveChangesID := store.AddAfterSaveChangesListener(afterSaveChanges)

	nBeforeDeleteCalledCount := 0
	beforeDelete := func(event *ravendb.BeforeDeleteEventArgs) {
		u, ok := event.Entity.(*User)
		assert.True(t, ok)
		assert.Equal(t, "users/1-A", u.ID)
		nBeforeDeleteCalledCount++
	}
	beforeDeleteID := store.AddBeforeDeleteListener(beforeDelete)

	nBeforeQueryCalledCount := 0
	beforeQuery := func(event *ravendb.BeforeQueryEventArgs) {
		nBeforeQueryCalledCount++
	}
	beforeQueryID := store.AddBeforeQueryListener(beforeQuery)

	{
		assert.Equal(t, 0, nBeforeStoreCalledCount)
		assert.Equal(t, 0, nAfterSaveChangesCalledCount)
		session := openSessionMust(t, store)
		users := goStore(t, session)
		session.Close()
		assert.Equal(t, len(users), nBeforeStoreCalledCount)
		assert.Equal(t, len(users), nAfterSaveChangesCalledCount)
	}

	{
		assert.Equal(t, 0, nBeforeDeleteCalledCount)
		session := openSessionMust(t, store)
		var u *User
		err = session.Load(&u, "users/1-A")
		assert.NoError(t, err)
		err = session.DeleteEntity(u)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
		assert.Equal(t, 1, nBeforeDeleteCalledCount)
	}

	{
		assert.Equal(t, 0, nBeforeQueryCalledCount)
		session := openSessionMust(t, store)
		var users []*User
		q := session.Query()
		err = q.GetResults(&users)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(users))
		session.Close()
		assert.Equal(t, 1, nBeforeQueryCalledCount)
	}

	store.RemoveBeforeStoreListener(beforeStoreID)
	store.RemoveAfterSaveChangesListener(afterSaveChangesID)
	store.RemoveBeforeDeleteListener(beforeDeleteID)
	store.RemoveBeforeQueryListener(beforeQueryID)

	{
		// verify those listeners were removed
		nBeforeStoreCalledCountPrev := nBeforeStoreCalledCount
		nAfterSaveChangesCalledCountPrev := nAfterSaveChangesCalledCount
		nBeforeDeleteCalledCountPrev := nBeforeDeleteCalledCount
		nBeforeQueryCalledCountPrev := nBeforeQueryCalledCount

		session := openSessionMust(t, store)

		var users []*User
		q := session.Query()
		err = q.GetResults(&users)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(users))
		assert.Equal(t, nBeforeQueryCalledCountPrev, nBeforeQueryCalledCount)

		u := &User{}
		err = session.Store(u)
		assert.NoError(t, err)
		err = session.Delete("users/2-A")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()

		assert.Equal(t, nBeforeStoreCalledCountPrev, nBeforeStoreCalledCount)
		assert.Equal(t, nAfterSaveChangesCalledCountPrev, nAfterSaveChangesCalledCount)
		assert.Equal(t, nBeforeDeleteCalledCountPrev, nBeforeDeleteCalledCount)
	}

	{
		// test that Refresh() only works if entity is in session
		session := openSessionMust(t, store)
		var u *User
		err = session.Load(&u, "users/3-A")
		assert.NoError(t, err)
		assert.NotNil(t, u)
		err = session.Refresh(u)
		assert.NoError(t, err)

		// TODO: refreshing the second time is invalid. This matches
		// Java behavior (I think) but I think it should be fine
		// The reason for that is that in refreshInternal() instead of
		// updating documentInfo.entity (which is the same object as entity
		// argument to Refresh()) we create a new object, set it as the
		// new documentInfo.entity and copy its properties to entity
		// to fix this, we should copy properties of the new object
		// to documentInfo.entity
		// https://github.com/ravendb/ravendb-go-client/issues/107
		err = session.Refresh(u)
		assertIllegalStateError(t, err, "Cannot refresh a transient instance")

		// test going over the limit of requests per session (32)
		// TODO: doesn't work because even Load() doesn't make second
		// Refresh() valid. Must fix https://github.com/ravendb/ravendb-go-client/issues/107
		// first
		/*
			n := 0
			for i := 0; err == nil && i < 32; i++ {
				u = nil
				err = session.Load(&u, "users/3-A")
				assert.NotNil(t, u)
				if err == nil {
					err = session.Refresh(u)
				}
				n++
			}
			assertIllegalStateError(t, err, "exceeded max number of requests per session of 32")
		*/

		session.Close()
	}

	{
		// check Load() does proper argument validation
		session := openSessionMust(t, store)

		var v *User
		err = session.Load(&v, "")
		assertIllegalArgumentError(t, err, "id cannot be empty string")

		err = session.Load(nil, "id")
		assertIllegalArgumentError(t, err, "result can't be nil")

		err = session.Load(User{}, "id")
		assertIllegalArgumentError(t, err, "result can't be of type tests.User, try passing **tests.User")

		err = session.Load(&User{}, "id")
		assertIllegalArgumentError(t, err, "result can't be of type *tests.User, try passing **tests.User")

		err = session.Load([]*User{}, "id")
		assertIllegalArgumentError(t, err, "result can't be of type []*tests.User")

		err = session.Load(&[]*User{}, "id")
		assertIllegalArgumentError(t, err, "result can't be of type *[]*tests.User")

		var n int
		err = session.Load(n, "id")
		assertIllegalArgumentError(t, err, "result can't be of type int")
		err = session.Load(&n, "id")
		assertIllegalArgumentError(t, err, "result can't be of type *int")
		nPtr := &n
		err = session.Load(&nPtr, "id")
		assertIllegalArgumentError(t, err, "result can't be of type **int")

		session.Close()
	}

}

func TestGo1(t *testing.T) {
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	go1Test(t, driver)
	goTestGetLastModifiedForAndChanges(t, driver)
	goTestListeners(t, driver)
}
