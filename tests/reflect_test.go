package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

type FooStruct struct {
	S string
	N int
}

func TestTypeName(t *testing.T) {
	v := FooStruct{}
	name := ravendb.GetFullTypeName(v)
	if name != "tests.FooStruct" {
		t.Fatalf("expected '%s', got '%s'", "ravendb.FooStruct", name)
	}
	name = ravendb.GetFullTypeName(&v)
	if name != "tests.FooStruct" {
		t.Fatalf("expected '%s', got '%s'", "ravendb.FooStruct", name)
	}
	name = ravendb.GetShortTypeNameName(v)
	if name != "FooStruct" {
		t.Fatalf("expected '%s', got '%s'", "FooStruct", name)
	}
	name = ravendb.GetShortTypeNameName(&v)
	if name != "FooStruct" {
		t.Fatalf("expected '%s', got '%s'", "FooStruct", name)
	}
}

func TestMakeStructFromJSONMap(t *testing.T) {
	s := &FooStruct{
		S: "str",
		N: 5,
	}
	jsmap := ravendb.StructToJSONMap(s)
	vd, err := json.Marshal(s)
	assert.NoError(t, err)
	typ := reflect.TypeOf(s)
	v2, err := ravendb.MakeStructFromJSONMap(typ, jsmap)
	assert.NoError(t, err)
	vTyp := fmt.Sprintf("%T", s)
	v2Typ := fmt.Sprintf("%T", v2)
	if vTyp != v2Typ {
		t.Fatalf("'%s' != '%s'", vTyp, v2Typ)
	}
	v2d, err := json.Marshal(v2)
	if !bytes.Equal(vd, v2d) {
		t.Fatalf("'%s' != '%s'", string(vd), string(v2d))
	}

	s2 := v2.(*FooStruct)
	assert.Equal(t, s.S, s2.S)
	assert.Equal(t, s.N, s2.N)
}

func TestIsStructy(t *testing.T) {
	v := FooStruct{}
	typ, ok := ravendb.GetStructTypeOfValue(v)
	if !ok || typ.Kind() != reflect.Struct {
		t.Fatalf("GetStructTypeOfValue(%T) returned false", v)
	}
	typ, ok = ravendb.GetStructTypeOfValue(&v)
	if !ok || typ.Kind() != reflect.Struct {
		t.Fatalf("GetStructTypeOfValue(%T) returned false", v)
	}
	v2 := "str"
	typ, ok = ravendb.GetStructTypeOfValue(v2)
	if ok {
		t.Fatalf("GetStructTypeOfValue(%T) returned true", v2)
	}
}

func TestGetIdentityProperty(t *testing.T) {
	got := ravendb.GetIdentityProperty(reflect.TypeOf(""))
	assert.Equal(t, "", got)
	got = ravendb.GetIdentityProperty(reflect.TypeOf(User{}))
	assert.Equal(t, "ID", got)
	got = ravendb.GetIdentityProperty(reflect.TypeOf(&User{}))
	assert.Equal(t, "ID", got)

	{
		// field not named ID
		v := struct {
			Id string
		}{}
		got = ravendb.GetIdentityProperty(reflect.TypeOf(v))
		assert.Equal(t, "", got)
	}

	{
		// field named ID but not stringa
		v := struct {
			ID int
		}{}
		got = ravendb.GetIdentityProperty(reflect.TypeOf(v))
		assert.Equal(t, "", got)
	}

}
