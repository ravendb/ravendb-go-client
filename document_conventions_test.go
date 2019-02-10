package ravendb

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type User struct {
	ID   string
	Name string
}

func TestDefaultGetCollectionName(t *testing.T) {
	// t.Parallel()

	name := getCollectionNameForTypeOrEntity(&User{})
	assert.Equal(t, "Users", name)
	name = getCollectionNameForTypeOrEntity(reflect.TypeOf(&User{}))
	assert.Equal(t, "Users", name)
}
