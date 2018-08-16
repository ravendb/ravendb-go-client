package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/ravendb/ravendb-go-client"
)

func TestDefaultGetCollectionName(t *testing.T) {
	name := ravendb.GefaultGetCollectionName(&User{})
	assert.Equal(t, "Users", name)
	name = ravendb.GefaultGetCollectionName(ravendb.GetTypeOf(&User{}))
	assert.Equal(t, "Users", name)
}
