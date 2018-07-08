package ravendb

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/jinzhu/copier"
)

// functionality related to reflection

// Go port of com.google.common.base.Defaults to make porting Java easier
func Defaults_defaultValue(clazz reflect.Type) interface{} {
	rv := reflect.Zero(clazz)
	return rv.Interface()
}

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

func getTypeOf(v interface{}) reflect.Type {
	// TODO: validate that v is of valid type (for now pointer to a struct)
	return reflect.TypeOf(v)
}

func isTypePrimitive(t reflect.Type) bool {
	kind := t.Kind()
	switch kind {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8,
		reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.String:
		return true
	case reflect.Ptr:
		return false
	// TODO: not all of those we should support
	case reflect.Array, reflect.Interface, reflect.Map, reflect.Slice, reflect.Struct:
		panicIf(true, "NYI")
	}
	return false
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

func isTypePointerToStruct(typ reflect.Type) bool {
	if typ.Kind() != reflect.Ptr {
		return false
	}
	typ = typ.Elem()
	return typ.Kind() == reflect.Struct
}

// given a json represented as map and type of a struct
func makeStructFromJSONMap(typ reflect.Type, js ObjectNode) (interface{}, error) {
	panicIf(!isTypePointerToStruct(typ), "typ should be pointer to struct but is %s, %s", typ.String(), typ.Kind().String())

	// reflect.New() creates a pointer to type. if typ is already a pointer,
	// we undo one level
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	rvNew := reflect.New(typ)
	d, err := json.Marshal(js)
	if err != nil {
		return nil, err
	}
	v := rvNew.Interface()
	err = json.Unmarshal(d, v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func dbglog(format string, args ...interface{}) string {
	s := fmt.Sprintf(format, args...)
	fmt.Println(s)
	return s
}

// corresponds to ObjectMapper.convertValue()
// val is coming from JSON, so it can be string, bool, float64, []interface{}
// or map[string]interface{}
// TODO: not sure about nil
// for simple types (int, bool, string) it should be just pass-through
// for structs decode ObjectNode => struct using makeStructFromJSONMap
func convertValue(val interface{}, clazz reflect.Type) (interface{}, error) {
	// TODO: implement every possible type. Need more comprehensive tests
	// to exercise those code paths
	switch clazz.Kind() {
	case reflect.String:
		switch v := val.(type) {
		case string:
			return v, nil
		default:
			panicIf(true, "%s", dbglog("converting of type %T to string NYI", val))
		}
	case reflect.Int:
		switch v := val.(type) {
		case int:
			return v, nil
		case float64:
			res := int(v)
			return res, nil
		default:
			panicIf(true, "%s", dbglog("converting of type %T to reflect.Int NYI", val))
		}
	case reflect.Ptr:
		clazz2 := clazz.Elem()
		switch clazz2.Kind() {
		case reflect.Struct:
			valIn, ok := val.(ObjectNode)
			if !ok {
				return nil, NewRavenException("can't convert value of type '%s' to a struct", val)
			}
			v, err := makeStructFromJSONMap(clazz, valIn)
			return v, err
		default:
			panicIf(true, "%s", dbglog("converting to pointer of '%s' NYI", clazz.Kind().String()))
		}
	default:
		panicIf(true, "%s", dbglog("converting to %s NYI", clazz.Kind().String()))
	}
	return nil, NewNotImplementedException("NYI")
}

// TODO: temporary name to match Java
// TODO: include github.com/jinzhu/copier to avoid dependency
func BeanUtils_copyProperties(dest Object, src Object) error {
	return copier.Copy(dest, src)
}
