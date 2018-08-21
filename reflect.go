package ravendb

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/jinzhu/copier"
)

// functionality related to reflection

// Go port of com.google.common.base.Defaults to make porting Java easier
func Defaults_defaultValue(clazz reflect.Type) interface{} {
	rv := reflect.Zero(clazz)
	return rv.Interface()
}

// GetFullTypeName returns fully qualified (including package) name of the type,
// after traversing pointers.
// e.g. for struct Foo in main package, the type of Foo and *Foo is main.Foo
func GetFullTypeName(v interface{}) string {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	typ := rv.Type()
	return typ.String()
}

// GetShortTypeNameName returns a short (not including package) name of the type,
// after traversing pointers.
// e.g. for struct Foo, the type of Foo and *Foo is "Foo"
// This is equivalent to Python's v.__class__.__name__
// Note: this emulates Java's operator over-loading to support
// GefaultGetCollectionName.
// We should have separate functions for reflect.Type and regular value
func GetShortTypeNameName(v interface{}) string {
	var typ reflect.Type
	var ok bool
	if typ, ok = v.(reflect.Type); ok {
		for typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}
	} else {
		rv := reflect.ValueOf(v)
		for rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		typ = rv.Type()
	}
	name := typ.Name()
	return name
}

// identity property is field of type string with name ID
func GetIdentityProperty(typ reflect.Type) string {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return ""
	}
	field, ok := typ.FieldByName("ID")
	if !ok || field.Type.Kind() != reflect.String {
		return ""
	}
	return "ID"
}

// GetTypeOf returns reflect.Type of a given value.
// TODO: possibly just call reflect.TypeOf directly
func GetTypeOf(v interface{}) reflect.Type {
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
		panic("NYI")
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

func GetStructTypeOfValue(v interface{}) (reflect.Type, bool) {
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


// if typ is ptr-to-struct, return as is
// if typ is ptr-to-ptr-to-struct, returns ptr-to-struct
// otherwise returns nil
func fixUpStructType(typ reflect.Type) reflect.Type {
	if typ.Kind() != reflect.Ptr {
		return nil
	}
	subtype := typ.Elem()
	if subtype.Kind() == reflect.Struct {
		return typ
	}
	if subtype.Kind() != reflect.Ptr {
		return nil
	}
	if subtype.Elem().Kind() == reflect.Struct {
		return subtype
	}
	return nil
}


func convertFloat64ToType(v float64, typ reflect.Type) interface{} {
	switch typ.Kind() {
	case reflect.Float32:
		return float32(v)
	case reflect.Float64:
		return v
	case reflect.Int:
		return int(v)
	case reflect.Int8:
		return int8(v)
	case reflect.Int16:
		return int16(v)
	case reflect.Int32:
		return int32(v)
	case reflect.Int64:
		return int64(v)
	case reflect.Uint:
		return uint(v)
	case reflect.Uint8:
		return uint8(v)
	case reflect.Uint16:
		return uint16(v)
	case reflect.Uint32:
		return uint32(v)
	case reflect.Uint64:
		return uint64(v)
	}
	panicIf(true, "don't know how to convert value of type %T to reflect type %s", v, typ.Name())
	return int(0)
}

func treeToValue(typ reflect.Type, js TreeNode) (interface{}, error) {
	// TODO: should also handle primitive types
	switch v := js.(type) {
	case string:
		if typ.Kind() == reflect.String {
			return js, nil
		}
		panicIf(true, "don't know how to convert value of type %T to reflect type %s", js, typ.Name())
	case float64:
		return convertFloat64ToType(v, typ), nil
	case bool:
		panicIf(true, "don't know how to convert value of type %T to reflect type %s", js, typ.Name())
	case []interface{}:
		panicIf(true, "don't know how to convert value of type %T to reflect type %s", js, typ.Name())
	case ObjectNode:
		return MakeStructFromJSONMap(typ, v)
	}
	panicIf(true, "don't know how to convert value of type %v to reflect type %s", js, typ.Name())
	return nil, fmt.Errorf("don't know how to convert value of type %v to reflect type %s", js, typ.Name())
}

// returns names of struct fields as serialized to JSON
func getJSONStructFieldNames(typ reflect.Type) []string {
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	panicIf(typ.Kind() != reflect.Struct, "we only support reflect.Struct, we got %s", typ.Kind())
	rvNew := reflect.New(typ)
	rv := reflect.ValueOf(rvNew)
	structType := rv.Type()
	var res []string
	for i := 0; i < rv.NumField(); i++ {
		structField := structType.Field(i)
		isExported := structField.PkgPath == ""
		if !isExported {
			continue
		}
		name := structField.Name
		tag := structField.Tag
		jsonTag := tag.Get("json")
		if jsonTag == "" {
			res = append(res, name)
			continue
		}
		// this could be "json,omitempty" etc. Extract just the name
		idx := strings.IndexByte(jsonTag, ',')
		if idx == -1 {
			res = append(res, jsonTag)
			continue
		}
		s := strings.TrimSpace(jsonTag[:idx-1])
		res = append(res, s)
	}

	return res
}

// given a json represented as map and type of a struct
func MakeStructFromJSONMap(typ reflect.Type, js ObjectNode) (interface{}, error) {
	if typ == GetTypeOf(ObjectNode{}) {
		return js, nil
	}
	typ2 := fixUpStructType(typ)
	panicIf(typ2 == nil, "typ should be pointer-to-struct or pointer-to-pointer-to-struct but is %s, %s", typ.String(), typ.Kind().String())

	typ = typ2
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
// for structs decode ObjectNode => struct using MakeStructFromJSONMap
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
			v, err := MakeStructFromJSONMap(clazz, valIn)
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
