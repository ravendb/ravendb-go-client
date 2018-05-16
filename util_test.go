package ravendb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

type FooStruct struct {
	S string
	N int
}

func TestTypeName(t *testing.T) {
	v := FooStruct{}
	name := getTypeName(v)
	if name != "ravendb.FooStruct" {
		t.Fatalf("expected '%s', got '%s'", "ravendb.FooStruct", name)
	}
	name = getTypeName(&v)
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
	v := &FooStruct{
		S: "str",
		N: 5,
	}
	jsmap := structToJSONMap(v)
	vd, err := json.Marshal(v)
	must(err)
	typ, ok := getStructTypeOfValue(v)
	if !ok {
		t.Fatalf("getStructTypeOfValue(%T) returned false", v)
	}
	v2 := makeStructFromJSONMap(typ, jsmap)
	vTyp := fmt.Sprintf("%T", v)
	v2Typ := fmt.Sprintf("%T", v2)
	if vTyp != v2Typ {
		t.Fatalf("'%s' != '%s'", vTyp, v2Typ)
	}
	v2d, err := json.Marshal(v2)
	if !bytes.Equal(vd, v2d) {
		t.Fatalf("'%s' != '%s'", string(vd), string(v2d))
	}
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
