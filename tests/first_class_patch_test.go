package tests

import (
	"testing"
	"time"
)

const (
	_docId = "users/1-A"
)

// Note: conflicts with User in user_test.go
type User2 struct {
	Stuff     []*Stuff  `json:"stuff"`
	LastLogin time.Time `json:"lastLogin"`
	Numbers   []int     `json:"numbers"`
}

type Stuff struct {
	Key    int               `json:"key"`
	Phone  string            `json:"phone"`
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

func firstClassPatch_canPatch(t *testing.T) {
	/*
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
			session.Advanced().Patch(_docId, "numbers[0]", 31)
			session.Advanced().Patch(_docId, "lastLogin", now)
			session.SaveChanges()

			session.Close()
		}
	*/

	// TODO: port this

	/*
	   try (IDocumentSession session = store.openSession()) {
	       User loaded = session.load(User.class, _docId);
	       assertThat(loaded.getNumbers()[0])
	               .isEqualTo(31);
	       assertThat(loaded.getLastLogin())
	               .isEqualTo(now);

	       session.advanced().patch(loaded, "stuff[0].phone", "123456");
	       session.saveChanges();
	   }

	   try (IDocumentSession session = store.openSession()) {
	       User loaded = session.load(User.class, _docId);
	       assertThat(loaded.getStuff()[0].getPhone())
	               .isEqualTo("123456");
	   }
	*/
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
	firstClassPatch_canPatch(t)
	firstClassPatch_canPatchAndModify(t)
	firstClassPatch_canPatchComplex(t)
}
