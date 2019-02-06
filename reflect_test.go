package ravendb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

type FooStruct struct {
	S   string
	N   int
	Foo *FooStruct
}

func TestDeserializeIncompatible(t *testing.T) {
	// unlike built-in json deserialization, we are forgiving of type mismatches
	js := `
{
	"S": "foo",
	"Foo": "foos/1"
}`
	var doc map[string]interface{}
	err := json.Unmarshal([]byte(js), &doc)
	assert.NoError(t, err)
	var fs *FooStruct
	err = makeStructFromJSONMap3(&fs, doc)
	assert.NoError(t, err)
}

func TestTypeName(t *testing.T) {
	v := FooStruct{}
	name := GetFullTypeName(v)
	assert.Equal(t, "ravendb.FooStruct", name)
	name = GetFullTypeName(&v)
	assert.Equal(t, "ravendb.FooStruct", name)
	name = getShortTypeNameName(v)
	assert.Equal(t, "FooStruct", name)
	name = getShortTypeNameName(&v)
	assert.Equal(t, "FooStruct", name)
}

func TestMakeStructFromJSONMap(t *testing.T) {
	s := &FooStruct{
		S: "str",
		N: 5,
	}
	jsmap := StructToJSONMap(s)
	vd, err := jsonMarshal(s)
	assert.NoError(t, err)
	typ := reflect.TypeOf(s)
	v2, err := makeStructFromJSONMap(typ, jsmap)

	assert.NoError(t, err)
	vTyp := fmt.Sprintf("%T", s)
	v2Typ := fmt.Sprintf("%T", v2)
	assert.Equal(t, vTyp, v2Typ)
	v2d, err := jsonMarshal(v2)
	assert.NoError(t, err)
	if !bytes.Equal(vd, v2d) {
		t.Fatalf("'%s' != '%s'", string(vd), string(v2d))
	}

	{
		s2 := v2.(*FooStruct)
		assert.Equal(t, s.S, s2.S)
		assert.Equal(t, s.N, s2.N)
	}

	var s2 *FooStruct
	err = makeStructFromJSONMap2(&s2, jsmap)
	assert.NoError(t, err)
	assert.Equal(t, s, s2)
}

func TestIsStructy(t *testing.T) {
	v := FooStruct{}
	typ, ok := getStructTypeOfValue(v)
	assert.True(t, ok && typ.Kind() == reflect.Struct)
	typ, ok = getStructTypeOfValue(&v)
	assert.True(t, ok && typ.Kind() == reflect.Struct)
	v2 := "str"
	_, ok = getStructTypeOfValue(v2)
	assert.False(t, ok)
}

func TestIsMapStringToPtrStruct(t *testing.T) {
	{
		v := map[string]*User{}
		tp, ok := isMapStringToPtrStruct(reflect.TypeOf(v))
		assert.True(t, ok)
		assert.Equal(t, reflect.TypeOf(&User{}), tp)
	}
	vals := []interface{}{
		1, true, 3.8, "string", []*User{}, map[string]User{}, map[int]*User{}, User{}, &User{},
	}
	for _, v := range vals {
		_, ok := isMapStringToPtrStruct(reflect.TypeOf(v))
		assert.False(t, ok)
	}
}

func TestGetIdentityProperty(t *testing.T) {
	got := getIdentityProperty(reflect.TypeOf(""))
	assert.Equal(t, "", got)
	got = getIdentityProperty(reflect.TypeOf(User{}))
	assert.Equal(t, "ID", got)
	got = getIdentityProperty(reflect.TypeOf(&User{}))
	assert.Equal(t, "ID", got)

	{
		// field not named ID
		v := struct {
			Id string
		}{}
		got = getIdentityProperty(reflect.TypeOf(v))
		assert.Equal(t, "", got)
	}

	{
		// field named ID but not stringa
		v := struct {
			ID int
		}{}
		got = getIdentityProperty(reflect.TypeOf(v))
		assert.Equal(t, "", got)
	}

}
