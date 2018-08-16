package tests

import (
	"testing"
	"time"

	ravendb "github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

// Note: conflicts with User in user_test.go
type User2 struct {
	Stuff     []*Stuff  `json:"stuff"`
	LastLogin *ravendb.ServerTime `json:"lastLogin"`
	Numbers   []int     `json:"numbers"`
}

type Stuff struct {
	Key    int               `json:"key"`
	Phone  *string            `json:"phone"`
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

func firstClassPatch_canPatch(t *testing.T) {
	stuff := []*Stuff{nil, nil, nil}
	stuff[0] = &Stuff{}
	stuff[0].Key = 6

	user := &User2{}
	user.Numbers = []int{66}
	user.Stuff = stuff

	var err error
	store := getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		err = session.Store(user)
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
		err = session.Advanced().PatchByID(_docId, "lastLogin", now)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		loadedI, err := session.Load(ravendb.GetTypeOf(&User2{}), _docId)
		assert.NoError(t, err)
		loaded := loadedI.(*User2)
		assert.Equal(t, loaded.Numbers[0], 31)
		assert.Equal(t, loaded.LastLogin, now)

		// TODO: this generates incorrect Script
		err = session.Advanced().PatchEntity(loaded, "stuff[0].phone", "123456")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)
		loadedI, err := session.Load(ravendb.GetTypeOf(&User2{}), _docId)
		assert.NoError(t, err)
		loaded := loadedI.(*User2)

		assert.Equal(t, loaded.Stuff[0].Phone, "123456")

		session.Close()
	}
}

func firstClassPatch_canPatchAndModify(t *testing.T)     {}
func firstClassPatch_canPatchComplex(t *testing.T)       {}
func firstClassPatch_canAddToArray(t *testing.T)         {}
func firstClassPatch_canRemoveFromArray(t *testing.T)    {}
func firstClassPatch_canIncrement(t *testing.T)          {}
func firstClassPatch_shouldMergePatchCalls(t *testing.T) {}

func TestFirstClassPatch(t *testing.T) {
	if dbTestsDisabled() {
		return
	}

	destroyDriver := createTestDriver(t)
	defer recoverTest(t, destroyDriver)

	// matches order of Java tests
	firstClassPatch_canIncrement(t)
	firstClassPatch_canAddToArray(t)
	firstClassPatch_canRemoveFromArray(t)
	firstClassPatch_shouldMergePatchCalls(t)
	if ravendb.EnableFailingTests {
		firstClassPatch_canPatch(t)
	}
	firstClassPatch_canPatchAndModify(t)
	firstClassPatch_canPatchComplex(t)
}
