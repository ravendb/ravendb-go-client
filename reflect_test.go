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
	S string
	N int
}

func TestTypeName(t *testing.T) {
	v := FooStruct{}
	name := GetFullTypeName(v)
	assert.Equal(t, "ravendb.FooStruct", name)
	name = GetFullTypeName(&v)
	assert.Equal(t, "ravendb.FooStruct", name)
	name = GetShortTypeNameName(v)
	assert.Equal(t, "FooStruct", name)
	name = GetShortTypeNameName(&v)
	assert.Equal(t, "FooStruct", name)
}

func TestMakeStructFromJSONMap(t *testing.T) {
	s := &FooStruct{
		S: "str",
		N: 5,
	}
	jsmap := StructToJSONMap(s)
	vd, err := json.Marshal(s)
	assert.NoError(t, err)
	typ := reflect.TypeOf(s)
	v2, err := MakeStructFromJSONMap(typ, jsmap)
	assert.NoError(t, err)
	vTyp := fmt.Sprintf("%T", s)
	v2Typ := fmt.Sprintf("%T", v2)
	assert.Equal(t, vTyp, v2Typ)
	v2d, err := json.Marshal(v2)
	assert.NoError(t, err)
	if !bytes.Equal(vd, v2d) {
		t.Fatalf("'%s' != '%s'", string(vd), string(v2d))
	}

	s2 := v2.(*FooStruct)
	assert.Equal(t, s.S, s2.S)
	assert.Equal(t, s.N, s2.N)
}

func TestIsStructy(t *testing.T) {
	v := FooStruct{}
	typ, ok := GetStructTypeOfValue(v)
	assert.True(t, ok && typ.Kind() == reflect.Struct)
	typ, ok = GetStructTypeOfValue(&v)
	assert.True(t, ok && typ.Kind() == reflect.Struct)
	v2 := "str"
	typ, ok = GetStructTypeOfValue(v2)
	assert.False(t, ok)
}

func TestGetIdentityProperty(t *testing.T) {
	got := GetIdentityProperty(reflect.TypeOf(""))
	assert.Equal(t, "", got)
	got = GetIdentityProperty(reflect.TypeOf(User{}))
	assert.Equal(t, "ID", got)
	got = GetIdentityProperty(reflect.TypeOf(&User{}))
	assert.Equal(t, "ID", got)

	{
		// field not named ID
		v := struct {
			Id string
		}{}
		got = GetIdentityProperty(reflect.TypeOf(v))
		assert.Equal(t, "", got)
	}

	{
		// field named ID but not stringa
		v := struct {
			ID int
		}{}
		got = GetIdentityProperty(reflect.TypeOf(v))
		assert.Equal(t, "", got)
	}

}
