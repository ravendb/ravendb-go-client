package tests

import (
	"fmt"
	"testing"
	"time"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

// Note: conflicts with User in user_test.go
type User2 struct {
	Stuff     []*Stuff     `json:"stuff"`
	LastLogin ravendb.Time `json:"lastLogin"`
	Numbers   []int        `json:"numbers"`
}

type Stuff struct {
	Key    int               `json:"key"`
	Phone  *string           `json:"phone"`
	Pet    *Pet              `json:"pet"`
	Friend *Friend           `json:"friend"`
	Dic    map[string]string `json:"dic"`
}

type Friend struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	Pet  *Pet   `json:"pet"`
}

type Pet struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
}

const (
	_docId = "user2s/1-A"
)

func firstClassPatch_canPatch(t *testing.T, driver *RavenTestDriver) {
	stuff := []*Stuff{nil, nil, nil}
	stuff[0] = &Stuff{}
	stuff[0].Key = 6

	user := &User2{}
	user.Numbers = []int{66}
	user.Stuff = stuff

	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		err = session.Store(user)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	now := ravendb.Time(time.Now().UTC())
	now = addDays(now, 2)
	{
		session := openSessionMust(t, store)

		err = session.Advanced().PatchByID(_docId, "numbers[0]", 31)
		assert.NoError(t, err)
		err = session.Advanced().PatchByID(_docId, "lastLogin", now)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var loaded *User2
		err = session.Load(&loaded, _docId)
		assert.NoError(t, err)
		assert.Equal(t, loaded.Numbers[0], 31)

		ll := time.Time(loaded.LastLogin)
		now2 := time.Time(now)
		if !ll.Equal(now2) {
			fmt.Printf("loaded.LastLogin: %s, now: %s\n", ll, now2)
		}
		assert.Equal(t, loaded.LastLogin, now)

		err = session.Advanced().PatchEntity(&loaded, "stuff[0].phone", "123456")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)
		var loaded *User2
		err = session.Load(&loaded, _docId)
		assert.NoError(t, err)

		assert.Equal(t, *loaded.Stuff[0].Phone, "123456")

		session.Close()
	}
}

func firstClassPatch_canPatchAndModify(t *testing.T, driver *RavenTestDriver) {
	user := &User2{}
	user.Numbers = []int{66}

	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		err = session.Store(user)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)
		var loaded *User2
		err = session.Load(&loaded, _docId)
		assert.NoError(t, err)
		loaded.Numbers[0] = 1

		err = session.Advanced().PatchEntity(&loaded, "numbers[0]", 2)
		assert.NoError(t, err)
		err = session.SaveChanges()
		_ = err.(*ravendb.IllegalStateException)

		session.Close()
	}
}

func firstClassPatch_canPatchComplex(t *testing.T, driver *RavenTestDriver) {
	stuff := []*Stuff{nil, nil, nil}
	stuff[0] = &Stuff{
		Key: 6,
	}

	user := &User2{
		Stuff: stuff,
	}

	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		err = session.Store(user)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		phone := "9255864406"
		newStuff := &Stuff{
			Key:    4,
			Phone:  &phone,
			Friend: &Friend{},
		}
		err = session.Advanced().PatchByID(_docId, "stuff[1]", newStuff)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var loaded *User2
		err = session.Load(&loaded, _docId)
		assert.NoError(t, err)
		assert.Equal(t, *loaded.Stuff[1].Phone, "9255864406")

		assert.Equal(t, loaded.Stuff[1].Key, 4)
		assert.NotNil(t, loaded.Stuff[1].Friend)

		pet1 := &Pet{
			Kind: "Dog",
			Name: "Hanan",
		}

		friendsPet := &Pet{
			Name: "Miriam",
			Kind: "Cat",
		}

		friend := &Friend{
			Name: "Gonras",
			Age:  28,
			Pet:  friendsPet,
		}

		phone := "9255864406"
		secondStuff := &Stuff{
			Key:    4,
			Phone:  &phone,
			Pet:    pet1,
			Friend: friend,
		}

		m := map[string]string{
			"Ohio":       "Columbus",
			"Utah":       "Salt Lake City",
			"Texas":      "Austin",
			"California": "Sacramento",
		}

		secondStuff.Dic = m

		err = session.Advanced().PatchEntity(&loaded, "stuff[2]", secondStuff)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var loaded *User2
		err = session.Load(&loaded, _docId)
		assert.NoError(t, err)

		assert.Equal(t, loaded.Stuff[2].Pet.Name, "Hanan")
		assert.Equal(t, loaded.Stuff[2].Friend.Name, "Gonras")
		assert.Equal(t, loaded.Stuff[2].Friend.Pet.Name, "Miriam")
		assert.Equal(t, len(loaded.Stuff[2].Dic), 4)
		assert.Equal(t, loaded.Stuff[2].Dic["Utah"], "Salt Lake City")

		session.Close()
	}
}

func firstClassPatch_canAddToArray(t *testing.T, driver *RavenTestDriver) {
	stuff := []*Stuff{nil}

	stuff[0] = &Stuff{}
	stuff[0].Key = 6

	user := &User2{
		Stuff:   stuff,
		Numbers: []int{1, 2},
	}

	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		err = session.Store(user)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)
		//push
		adder := func(roles *ravendb.JavaScriptArray) {
			roles.Add(3)
		}
		err = session.Advanced().PatchArrayByID(_docId, "numbers", adder)
		assert.NoError(t, err)

		adder = func(roles *ravendb.JavaScriptArray) {
			stuff1 := &Stuff{
				Key: 75,
			}
			roles.Add(stuff1)
		}
		err = session.Advanced().PatchArrayByID(_docId, "stuff", adder)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var loaded *User2
		err = session.Load(&loaded, _docId)
		assert.NoError(t, err)

		assert.Equal(t, loaded.Numbers[2], 3)
		assert.Equal(t, loaded.Stuff[1].Key, 75)

		adder := func(roles *ravendb.JavaScriptArray) {
			roles.Add(101, 102, 103)
		}
		err = session.Advanced().PatchArrayInEntity(&loaded, "numbers", adder)
		assert.NoError(t, err)
		adder = func(roles *ravendb.JavaScriptArray) {
			s1 := &Stuff{
				Key: 102,
			}

			phone := "123456"
			s2 := &Stuff{
				Phone: &phone,
			}

			roles.Add(s1).Add(s2)
		}
		err = session.Advanced().PatchArrayInEntity(&loaded, "stuff", adder)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var loaded *User2
		err = session.Load(&loaded, _docId)
		assert.NoError(t, err)
		assert.Equal(t, len(loaded.Numbers), 6)

		assert.Equal(t, loaded.Numbers[5], 103)

		assert.Equal(t, loaded.Stuff[2].Key, 102)

		assert.Equal(t, *loaded.Stuff[3].Phone, "123456")

		adder := func(roles *ravendb.JavaScriptArray) {
			roles.Add(201, 202, 203)
		}

		err = session.Advanced().PatchArrayInEntity(&loaded, "numbers", adder)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var loaded *User2
		err = session.Load(&loaded, _docId)
		assert.NoError(t, err)

		assert.Equal(t, len(loaded.Numbers), 9)
		assert.Equal(t, loaded.Numbers[7], 202)

		session.Close()
	}
}

func firstClassPatch_canRemoveFromArray(t *testing.T, driver *RavenTestDriver) {
	stuff := []*Stuff{nil, nil}
	stuff[0] = &Stuff{
		Key: 6,
	}

	phone := "123456"
	stuff[1] = &Stuff{
		Phone: &phone,
	}

	user := &User2{
		Stuff:   stuff,
		Numbers: []int{1, 2, 3},
	}

	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		err = session.Store(user)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		adder := func(roles *ravendb.JavaScriptArray) {
			roles.RemoveAt(1)
		}
		err = session.Advanced().PatchArrayByID(_docId, "numbers", adder)
		assert.NoError(t, err)
		adder = func(roles *ravendb.JavaScriptArray) {
			roles.RemoveAt(0)
		}
		err = session.Advanced().PatchArrayByID(_docId, "stuff", adder)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		var loaded *User2
		err = session.Load(&loaded, _docId)
		assert.NoError(t, err)
		assert.Equal(t, len(loaded.Numbers), 2)
		assert.Equal(t, loaded.Numbers[1], 3)

		assert.Equal(t, len(loaded.Stuff), 1)
		assert.Equal(t, *loaded.Stuff[0].Phone, "123456")

		session.Close()
	}
}

func firstClassPatch_canIncrement(t *testing.T, driver *RavenTestDriver) {
	s := []*Stuff{nil, nil, nil}
	s[0] = &Stuff{
		Key: 6,
	}

	user := &User2{
		Numbers: []int{66},
		Stuff:   s,
	}

	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		err = session.Store(user)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		err = session.Advanced().IncrementByID(_docId, "numbers[0]", 1)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)
		var loaded *User2
		err = session.Load(&loaded, _docId)
		assert.NoError(t, err)
		assert.Equal(t, loaded.Numbers[0], 67)

		err = session.Advanced().IncrementEntity(&loaded, "stuff[0].key", -3)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)
		var loaded *User2
		err = session.Load(&loaded, _docId)
		assert.NoError(t, err)

		assert.Equal(t, loaded.Stuff[0].Key, 3)
		session.Close()
	}
}

func firstClassPatch_shouldMergePatchCalls(t *testing.T, driver *RavenTestDriver) {
	stuff := []*Stuff{nil, nil, nil}
	stuff[0] = &Stuff{
		Key: 6,
	}

	user := &User2{
		Stuff:   stuff,
		Numbers: []int{66},
	}

	user2 := &User2{
		Numbers: []int{1, 2, 3},
		Stuff:   stuff,
	}

	docId2 := "user2s/2-A"

	var err error
	store := getDocumentStoreMust(t, driver)
	defer store.Close()

	{
		session := openSessionMust(t, store)

		err = session.Store(user)
		assert.NoError(t, err)
		err = session.StoreWithID(user2, docId2)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	now := time.Now()

	{
		session := openSessionMust(t, store)

		err = session.Advanced().PatchByID(_docId, "numbers[0]", 31)
		assert.NoError(t, err)
		assert.Equal(t, session.GetDeferredCommandsCount(), 1)

		err = session.Advanced().PatchByID(_docId, "lastLogin", now)
		assert.NoError(t, err)
		assert.Equal(t, session.GetDeferredCommandsCount(), 1)

		err = session.Advanced().PatchByID(docId2, "numbers[0]", 123)
		assert.NoError(t, err)
		assert.Equal(t, session.GetDeferredCommandsCount(), 2)

		err = session.Advanced().PatchByID(docId2, "lastLogin", now)
		assert.NoError(t, err)
		assert.Equal(t, session.GetDeferredCommandsCount(), 2)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		err = session.Advanced().IncrementByID(_docId, "numbers[0]", 1)
		assert.NoError(t, err)
		assert.Equal(t, session.GetDeferredCommandsCount(), 1)

		adder := func(r *ravendb.JavaScriptArray) {
			r.Add(77)
		}
		err = session.Advanced().PatchArrayByID(_docId, "numbers", adder)
		assert.NoError(t, err)
		assert.Equal(t, session.GetDeferredCommandsCount(), 1)

		adder = func(r *ravendb.JavaScriptArray) {
			r.Add(88)
		}
		err = session.Advanced().PatchArrayByID(_docId, "numbers", adder)
		assert.NoError(t, err)
		assert.Equal(t, session.GetDeferredCommandsCount(), 1)

		adder = func(r *ravendb.JavaScriptArray) {
			r.RemoveAt(1)
		}
		err = session.Advanced().PatchArrayByID(_docId, "numbers", adder)
		assert.NoError(t, err)
		assert.Equal(t, session.GetDeferredCommandsCount(), 1)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}
}

func TestFirstClassPatch(t *testing.T) {
	if dbTestsDisabled() {
		return
	}
	t.Parallel()

	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	// matches order of Java tests
	firstClassPatch_canIncrement(t, driver)
	firstClassPatch_canAddToArray(t, driver)
	firstClassPatch_canRemoveFromArray(t, driver)
	firstClassPatch_shouldMergePatchCalls(t, driver)
	firstClassPatch_canPatch(t, driver)
	firstClassPatch_canPatchAndModify(t, driver)
	firstClassPatch_canPatchComplex(t, driver)
}
