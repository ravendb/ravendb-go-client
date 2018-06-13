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
	name := getFullTypeName(v)
	if name != "ravendb.FooStruct" {
		t.Fatalf("expected '%s', got '%s'", "ravendb.FooStruct", name)
	}
	name = getFullTypeName(&v)
	if name != "ravendb.FooStruct" {
		t.Fatalf("expected '%s', got '%s'", "ravendb.FooStruct", name)
	}
	name = getShortTypeName(v)
	if name != "FooStruct" {
		t.Fatalf("expected '%s', got '%s'", "FooStruct", name)
	}
	name = getShortTypeName(&v)
	if name != "FooStruct" {
		t.Fatalf("expected '%s', got '%s'", "FooStruct", name)
	}
}

func TestMakeStructFromJSONMap(t *testing.T) {
	s := &FooStruct{
		S: "str",
		N: 5,
	}
	jsmap := structToJSONMap(s)
	vd, err := json.Marshal(s)
	assert.NoError(t, err)
	typ := getTypeOfValue(s)
	v2, err := makeStructFromJSONMap(typ, jsmap)
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
	typ, ok := getStructTypeOfValue(v)
	if !ok || typ.Kind() != reflect.Struct {
		t.Fatalf("getStructTypeOfValue(%T) returned false", v)
	}
	typ, ok = getStructTypeOfValue(&v)
	if !ok || typ.Kind() != reflect.Struct {
		t.Fatalf("getStructTypeOfValue(%T) returned false", v)
	}
	v2 := "str"
	typ, ok = getStructTypeOfValue(v2)
	if ok {
		t.Fatalf("getStructTypeOfValue(%T) returned true", v2)
	}
}
