package ravendb

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

type FooStruct struct {
	S    string
	SPtr *string
	N    int
	Foo  *FooStruct
}

func TestDeserializeIncompatible(t *testing.T) {
	// unlike built-in json deserialization, we are forgiving of type mismatches
	js := `
{
	"S": "foo",
	"N": 3.3,
	"Foo": "foos/1"
}`
	var doc map[string]interface{}
	err := json.Unmarshal([]byte(js), &doc)
	assert.NoError(t, err)
	var fs *FooStruct
	err = makeStructFromJSONMap3(&fs, doc)
	assert.NoError(t, err)
	assert.Equal(t, fs.S, "foo")
}
