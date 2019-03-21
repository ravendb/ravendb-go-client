package tests

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	ravendb "github.com/ravendb/ravendb-go-client"

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
		err = session.Delete(v)
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
		err = session.Advanced().Patch(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().Increment(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().PatchArray(v, "foo", nil)
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
		err = session.Delete(v)
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
		err = session.Advanced().Patch(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().Increment(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().PatchArray(v, "foo", nil)
		assertIllegalArgumentError(t, err)
		err = session.Refresh(v)
		assertIllegalArgumentError(t, err)
	}

	{
		// can't store/delete etc. struct
		v := user
		err = session.Store(v)
		assertIllegalArgumentError(t, err, "entity can't be of type tests.User, try passing *tests.User")
		err = session.StoreWithID(v, "users/1")
		assertIllegalArgumentError(t, err)
		err = session.Delete(v)
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
		err = session.Advanced().Patch(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().Increment(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().PatchArray(v, "foo", nil)
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
		err = session.Delete(v)
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
		err = session.Advanced().Patch(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().Increment(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().PatchArray(v, "foo", nil)
		assertIllegalArgumentError(t, err)
		err = session.Refresh(v)
		assertIllegalArgumentError(t, err)
	}

	{
		// can't store/delete etc. a map
		var v map[string]interface{}
		err = session.Store(v)
		assertIllegalArgumentError(t, err, "entity can't be of type map[string]interface {}, try passing *map[string]interface {}")
		err = session.StoreWithID(v, "users/1")
		assertIllegalArgumentError(t, err)
		err = session.Delete(v)
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
		err = session.Advanced().Patch(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().Increment(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().PatchArray(v, "foo", nil)
		assertIllegalArgumentError(t, err)
		err = session.Refresh(v)
		assertIllegalArgumentError(t, err)
	}

	{
		v := &User{} // dummy value that only has to pass type check
		adv := session.Advanced()

		err = adv.Increment(v, "", 1)
		assertIllegalArgumentError(t, err, "path can't be empty string")
		err = adv.Increment(v, "foo", nil)
		assertIllegalArgumentError(t, err, "valueToAdd can't be nil")

		err = adv.IncrementByID("", "foo", 1)
		assertIllegalArgumentError(t, err, "id can't be empty string")
		err = adv.IncrementByID("id", "", 1)
		assertIllegalArgumentError(t, err, "path can't be empty string")
		err = adv.IncrementByID("id", "foo", nil)
		assertIllegalArgumentError(t, err, "valueToAdd can't be nil")

		err = adv.Patch(v, "", 1)
		assertIllegalArgumentError(t, err, "path can't be empty string")
		err = adv.Patch(v, "foo", nil)
		assertIllegalArgumentError(t, err, "value can't be nil")

		err = adv.PatchByID("", "foo", 1)
		assertIllegalArgumentError(t, err, "id can't be empty string")
		err = adv.PatchByID("id", "", 1)
		assertIllegalArgumentError(t, err, "path can't be empty string")
		err = adv.PatchByID("id", "foo", nil)
		assertIllegalArgumentError(t, err, "value can't be nil")

		err = adv.PatchArray(v, "", nil)
		assertIllegalArgumentError(t, err, "pathToArray can't be empty string")
		err = adv.PatchArray(v, "foo", nil)
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

		err = session.Delete(u)
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
		err = session.Delete(u)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
		assert.Equal(t, 1, nBeforeDeleteCalledCount)
	}

	{
		assert.Equal(t, 0, nBeforeQueryCalledCount)
		session := openSessionMust(t, store)
		tp := reflect.TypeOf(&User{})
		q := session.QueryCollectionForType(tp)
		var users []*User
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
		q := session.QueryCollectionForType(userType)
		err = q.GetResults(&users)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(users))
		assert.Equal(t, nBeforeQueryCalledCountPrev, nBeforeQueryCalledCount)

		u := &User{}
		err = session.Store(u)
		assert.NoError(t, err)
		err = session.DeleteByID("users/2-A", "")
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

		err = session.Refresh(u)
		assert.NoError(t, err)

		for i := 0; err == nil && i < 32; i++ {
			err = session.Refresh(u)
		}
		assertIllegalStateError(t, err, "exceeded max number of requests per session of 32")

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

	{
		// check LoadMulti() does proper argument validation
		session := openSessionMust(t, store)

		var v map[string]*User
		err = session.LoadMulti(v, nil)
		assertIllegalArgumentError(t, err, "ids cannot be empty array")
		err = session.LoadMulti(&v, []string{})
		assertIllegalArgumentError(t, err, "ids cannot be empty array")

		err = session.LoadMulti(User{}, []string{"id"})
		assertIllegalArgumentError(t, err, "results can't be of type tests.User, must be map[string]<type>")

		err = session.LoadMulti(&User{}, []string{"id"})
		assertIllegalArgumentError(t, err, "results can't be of type *tests.User, must be map[string]<type>")

		err = session.LoadMulti(map[int]*User{}, []string{"id"})
		assertIllegalArgumentError(t, err, "results can't be of type map[int]*tests.User, must be map[string]<type>")

		err = session.LoadMulti(map[string]int{}, []string{"id"})
		assertIllegalArgumentError(t, err, "results can't be of type map[string]int, must be map[string]<type>")

		err = session.LoadMulti(map[string]*int{}, []string{"id"})
		assertIllegalArgumentError(t, err, "results can't be of type map[string]*int, must be map[string]<type>")

		err = session.LoadMulti(map[string]User{}, []string{"id"})
		assertIllegalArgumentError(t, err, "results can't be of type map[string]tests.User, must be map[string]<type>")

		err = session.LoadMulti(v, []string{"id"})
		assertIllegalArgumentError(t, err, "results can't be a nil map")

		session.Close()
	}

}

// TODO: this must be more comprehensive. Need to test all APIs.
func goTestStoreMap(t *testing.T, driver *RavenTestDriver) {
	var err error

	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		m := map[string]interface{}{
			"foo":     5,
			"bar":     true,
			"nullVal": nil,
			"strVal":  "a string",
		}
		err = session.StoreWithID(&m, "maps/1")
		assert.NoError(t, err)

		m2 := map[string]interface{}{
			"foo":    8,
			"strVal": "more string",
		}
		err = session.Store(&m2)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		meta, err := session.GetMetadataFor(m)
		assertIllegalArgumentError(t, err, "instance can't be of type map[string]interface {}, try passing *map[string]interface {}")
		assert.Nil(t, meta)

		meta, err = session.GetMetadataFor(&m)
		assert.NoError(t, err)
		assert.NotNil(t, meta)

		session.Close()
	}

	{
		session := openSessionMust(t, store)
		var mp *map[string]interface{}
		err = session.Load(&mp, "maps/1")
		assert.NoError(t, err)
		m := *mp
		assert.Equal(t, float64(5), m["foo"])
		assert.Equal(t, "a string", m["strVal"])

		session.Close()
	}
}

func goTestFindCollectionName(t *testing.T) {
	findCollectionName := func(entity interface{}) string {
		if _, ok := entity.(*User); ok {
			return "my users"
		}
		return ravendb.GetCollectionNameDefault(entity)
	}
	c := ravendb.NewDocumentConventions()
	c.FindCollectionName = findCollectionName
	name := c.GetCollectionName(&Employee{})
	assert.Equal(t, name, "Employees")

	name = c.GetCollectionName(&User{})
	assert.Equal(t, name, "my users")
}

// test that insertion order of bulk_docs (BatchOperation / BatchCommand)
func goTestBatchCommandOrder(t *testing.T, driver *RavenTestDriver) {
	var err error

	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	// delete to trigger a code path that uses deferred commands
	// this is very sensitive to how code is structured: deleted
	// commands are gathered first in random order and put
	// commands are in insertion order
	nUsers := 10
	{
		session := openSessionMust(t, store)
		ids := []string{"users/5"}
		for i := 1; i <= nUsers; i++ {
			u := &User{
				Age: i,
			}
			u.setName(fmt.Sprintf("Name %d", i))
			id := fmt.Sprintf("users/%d", i)
			err = session.StoreWithID(u, id)
			assert.NoError(t, err)
			if i == 5 {
				err = session.Delete(u)
				assert.NoError(t, err)
			} else {
				ids = append(ids, id)
			}
		}
		commandsData, err := session.ForTestsSaveChangesGetCommands()
		assert.NoError(t, err)
		assert.Equal(t, len(commandsData), nUsers)
		for i, cmdData := range commandsData {
			var id string
			switch d := cmdData.(type) {
			case *ravendb.PutCommandDataWithJSON:
				id = d.ID
			case *ravendb.DeleteCommandData:
				id = d.ID
			}
			expID := ids[i]
			assert.Equal(t, expID, id)
			assert.Equal(t, expID, id)
		}
		session.Close()
	}
}

// test that we get a meaningful error for server exceptions sent as JSON response
// https://github.com/ravendb/ravendb-go-client/issues/147
func goTestInvalidIndexDefinition(t *testing.T, driver *RavenTestDriver) {
	var err error

	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	indexName := "Song/TextData"
	index := ravendb.NewIndexCreationTask(indexName)

	index.Map = `
from song in docs.Songs
select {
	SongData = new {
		song.Artist,
		song.Title,
		song.Tags,
		song.TrackId
	}
}
`
	index.Index("SongData", ravendb.FieldIndexingSearch)

	err = index.Execute(store, nil, "")
	assert.Error(t, err)
	_, ok := err.(*ravendb.IndexCompilationError)
	assert.True(t, ok)
}

func TestGo1(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	goTestStoreMap(t, driver)
	go1Test(t, driver)
	goTestGetLastModifiedForAndChanges(t, driver)
	goTestListeners(t, driver)
	goTestFindCollectionName(t)
	goTestBatchCommandOrder(t, driver)
	goTestInvalidIndexDefinition(t, driver)
}
