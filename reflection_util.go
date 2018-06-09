package ravendb

import (
	"encoding/json"
	"reflect"
)

// functionality related to reflection

// getFullTypeName returns fully qualified (including package) name of the type,
// after traversing pointers.
// e.g. for struct Foo in main package, the type of Foo and *Foo is main.Foo
func getFullTypeName(v interface{}) string {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	typ := rv.Type()
	return typ.String()
}

// getShortTypeName returns a short (not including package) name of the type,
// after traversing pointers.
// e.g. for struct Foo, the type of Foo and *Foo is "Foo"
// This is equivalent to Python's v.__class__.__name__
func getShortTypeName(v interface{}) string {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	typ := rv.Type()
	return typ.Name()
}

func getTypeOfValue(v interface{}) reflect.Type {
	// TODO: validate that v is of valid type (for now pointer to a struct)
	return reflect.TypeOf(v)
}

func getStructTypeOfReflectValue(rv reflect.Value) (reflect.Type, bool) {
	if rv.Type().Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	typ := rv.Type()
	if typ.Kind() == reflect.Struct {
		return typ, true
	}
	return typ, false
}

func getStructTypeOfValue(v interface{}) (reflect.Type, bool) {
	rv := reflect.ValueOf(v)
	return getStructTypeOfReflectValue(rv)
}

// given a json represented as map and type of a struct
func makeStructFromJSONMap(typ reflect.Type, js ObjectNode) interface{} {
	panicIf(typ.Kind() != reflect.Struct, "rv should be of type Struct but is %s", typ.String())
	rvNew := reflect.New(typ)
	d, err := json.Marshal(js)
	must(err)
	v := rvNew.Interface()
	err = json.Unmarshal(d, v)
	must(err)
	return v
}
