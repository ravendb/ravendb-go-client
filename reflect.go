package ravendb

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/jinzhu/copier"
)

// functionality related to reflection

func isPtrStruct(t reflect.Type) (reflect.Type, bool) {
	if t.Kind() == reflect.Ptr && t.Elem() != nil && t.Elem().Kind() == reflect.Struct {
		return t, true
	}
	return nil, false
}

func isPtrPtrStruct(tp reflect.Type) (reflect.Type, bool) {
	if tp.Kind() != reflect.Ptr {
		return nil, false
	}
	return isPtrStruct(tp.Elem())
}

func isPtrSlicePtrStruct(tp reflect.Type) (reflect.Type, bool) {
	if tp.Kind() != reflect.Ptr {
		return nil, false
	}
	tp = tp.Elem()
	if tp.Kind() != reflect.Slice {
		return nil, false
	}
	return isPtrStruct(tp.Elem())
}

func isMapStringToPtrStruct(tp reflect.Type) (reflect.Type, bool) {
	if tp.Kind() != reflect.Map {
		return nil, false
	}
	if tp.Key().Kind() != reflect.String {
		return nil, false
	}
	return isPtrStruct(tp.Elem())
}

// Go port of com.google.common.base.Defaults to make porting Java easier
func getDefaultValueForType(clazz reflect.Type) interface{} {
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
func getShortTypeNameName(v interface{}) string {
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
func getIdentityProperty(typ reflect.Type) string {
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

func getStructTypeOfValue(v interface{}) (reflect.Type, bool) {
	rv := reflect.ValueOf(v)
	return getStructTypeOfReflectValue(rv)
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

func treeToValue(typ reflect.Type, js interface{}) (interface{}, error) {
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
	case map[string]interface{}:
		return makeStructFromJSONMap(typ, v)
	}
	panicIf(true, "don't know how to convert value of type %v to reflect type %s", js, typ.Name())
	return nil, fmt.Errorf("don't know how to convert value of type %v to reflect type %s", js, typ.Name())
}

// get name of struct field for json serialization
// empty string means we should skip this field
func getJSONFieldName(field reflect.StructField) string {
	// skip unexported fields
	if field.PkgPath != "" {
		return ""
	}

	tag := field.Tag.Get("json")
	// if no tag, use field name
	if tag == "" {
		return field.Name
	}
	// skip if explicitly marked as non-json serializable
	// TODO: write tests for this
	if tag == "-" {
		return ""
	}
	// this could be "name,omitempty" etc.; extract just the name
	if idx := strings.IndexByte(tag, ','); idx != -1 {
		name := tag[:idx-1]
		// if it's sth. like ",omitempty", use field name
		// TODO: write tests for this
		if name == "" {
			return field.Name
		}
		return name
	}
	return tag
}

// FieldsFor returns names of all fields for the value of a struct type.
// They can be used in e.g. DocumentQuery.SelectFields:
// fields := ravendb.FieldsFor(&MyType{})
// q = q.SelectFields(fields...)
func FieldsFor(s interface{}) []string {
	v := reflect.ValueOf(s)
	// if pointer get the underlying element≤
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	panicIf(v.Kind() != reflect.Struct, "argument must be struct, we got %T", s)
	t := v.Type()
	var res []string
	for i := 0; i < t.NumField(); i++ {
		if name := getJSONFieldName(t.Field(i)); name != "" {
			res = append(res, name)
		}
	}
	return res
}

// given js value (most likely as map[string]interface{}) decode into res
func decodeJSONAsStruct(js interface{}, res interface{}) error {
	d, err := jsonMarshal(js)
	if err != nil {
		return err
	}
	return jsonUnmarshal(d, res)
}

// given a json represented as map and type of a struct
func makeStructFromJSONMap(typ reflect.Type, js map[string]interface{}) (interface{}, error) {
	if typ == reflect.TypeOf(map[string]interface{}{}) {
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
	d, err := jsonMarshal(js)
	if err != nil {
		return nil, err
	}
	v := rvNew.Interface()
	err = jsonUnmarshal(d, v)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// result should be *<type> and we'll do equivalent of: *result = <type>
func makeStructFromJSONMap2(result interface{}, js map[string]interface{}) error {
	// TODO: not sure if should accept result of *map[string]interface{} or map[string]interface{}
	if res, ok := result.(*map[string]interface{}); ok {
		*res = js
		return nil
	}
	if res, ok := result.(map[string]interface{}); ok {
		for k, v := range js {
			res[k] = v
		}
		return nil
	}

	d, err := jsonMarshal(js)
	if err != nil {
		return err
	}
	return jsonUnmarshal(d, result)
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
// for structs decode map[string]interface{} => struct using MakeStructFromJSONMap
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
			valIn, ok := val.(map[string]interface{})
			if !ok {
				return nil, newRavenError("can't convert value of type '%s' to a struct", val)
			}
			v, err := makeStructFromJSONMap(clazz, valIn)
			return v, err
		default:
			panicIf(true, "%s", dbglog("converting to pointer of '%s' NYI", clazz.Kind().String()))
		}
	default:
		panicIf(true, "%s", dbglog("converting to %s NYI", clazz.Kind().String()))
	}
	return nil, newNotImplementedError("NYI")
}

// TODO: include github.com/jinzhu/copier to avoid dependency
func copyValueProperties(dest interface{}, src interface{}) error {
	return copier.Copy(dest, src)
}

// m is a single-element map[string]*struct
// returns single map value
func getSingleMapValue(results interface{}) (interface{}, error) {
	m := reflect.ValueOf(results)
	if m.Type().Kind() != reflect.Map {
		return nil, fmt.Errorf("results should be a map[string]*struct, is %s. tp: %s", m.Type().String(), m.Type().String())
	}
	mapKeyType := m.Type().Key()
	if mapKeyType != stringType {
		return nil, fmt.Errorf("results should be a map[string]*struct, is %s. tp: %s", m.Type().String(), m.Type().String())
	}
	mapElemPtrType := m.Type().Elem()
	if mapElemPtrType.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("results should be a map[string]*struct, is %s. tp: %s", m.Type().String(), m.Type().String())
	}

	mapElemType := mapElemPtrType.Elem()
	if mapElemType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("results should be a map[string]*struct, is %s. tp: %s", m.Type().String(), m.Type().String())
	}
	keys := m.MapKeys()
	if len(keys) == 0 {
		return nil, nil
	}
	if len(keys) != 1 {
		return nil, fmt.Errorf("expected results to have only one element, has %d", len(keys))
	}
	v := m.MapIndex(keys[0])
	return v.Interface(), nil
}
