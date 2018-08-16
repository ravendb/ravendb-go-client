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

func firstClassPatch_canPatch(t *testing.T)              {}
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
