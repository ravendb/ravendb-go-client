package ravendb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultGetCollectionName(t *testing.T) {
	name := defaultGetCollectionName(&User{})
	assert.Equal(t, "Users", name)
	name = defaultGetCollectionName(GetTypeOf(&User{}))
	assert.Equal(t, "Users", name)
}
