package tests

import (
	"reflect"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func TestDefaultGetCollectionName(t *testing.T) {
	name := ravendb.DefaultGetCollectionName(&User{})
	assert.Equal(t, "Users", name)
	name = ravendb.DefaultGetCollectionName(reflect.TypeOf(&User{}))
	assert.Equal(t, "Users", name)
}
