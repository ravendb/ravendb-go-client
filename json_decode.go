package ravendb

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

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

// this is lifted from encoding\json\decode.go in std library

// indirect walks down v allocating pointers as needed,
// until it gets to a non-pointer.
// if it encounters an Unmarshaler, indirect stops and returns that.
// if decodingNull is true, indirect stops at the last pointer so it can be set to nil.
func indirect(v reflect.Value, decodingNull bool) (json.Unmarshaler, encoding.TextUnmarshaler, reflect.Value) {
	// Issue #24153 indicates that it is generally not a guaranteed property
	// that you may round-trip a reflect.Value by calling Value.Addr().Elem()
	// and expect the value to still be settable for values derived from
	// unexported embedded struct fields.
	//
	// The logic below effectively does this when it first addresses the value
	// (to satisfy possible pointer methods) and continues to dereference
	// subsequent pointers as necessary.
	//
	// After the first round-trip, we set v back to the original value to
	// preserve the original RW flags contained in reflect.Value.
	v0 := v
	haveAddr := false

	// If v is a named type and is addressable,
	// start with its address, so that if the type has pointer methods,
	// we find them.
	if v.Kind() != reflect.Ptr && v.Type().Name() != "" && v.CanAddr() {
		haveAddr = true
		v = v.Addr()
	}
	for {
		// Load value from interface, but only if the result will be
		// usefully addressable.
		if v.Kind() == reflect.Interface && !v.IsNil() {
			e := v.Elem()
			if e.Kind() == reflect.Ptr && !e.IsNil() && (!decodingNull || e.Elem().Kind() == reflect.Ptr) {
				haveAddr = false
				v = e
				continue
			}
		}

		if v.Kind() != reflect.Ptr {
			break
		}

		if v.Elem().Kind() != reflect.Ptr && decodingNull && v.CanSet() {
			break
		}
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		if v.Type().NumMethod() > 0 {
			if u, ok := v.Interface().(json.Unmarshaler); ok {
				return u, nil, reflect.Value{}
			}
			if !decodingNull {
				if u, ok := v.Interface().(encoding.TextUnmarshaler); ok {
					return nil, u, reflect.Value{}
				}
			}
		}

		if haveAddr {
			v = v0 // restore original value after round-trip Value.Addr().Elem()
			haveAddr = false
		} else {
			v = v.Elem()
		}
	}
	return nil, nil, v
}

// based on func (d *decodeState) object(v reflect.Value) error {
// in encoding\json\decode.go in std library
func jsonDecodeObject(v reflect.Value, m map[string]interface{}) error {
	// Check for unmarshaler.
	u, ut, pv := indirect(v, false)
	if u != nil {
		d, err := json.Marshal(m)
		if err != nil {
			return err
		}
		return u.UnmarshalJSON(d)
	}
	if ut != nil {
		return errors.New("json: cannot unmarshal object into Go value of type " + v.Type().String())
	}
	v = pv

	// Decoding into nil interface? Switch to non-reflect code.
	if v.Kind() == reflect.Interface && v.NumMethod() == 0 {
		var oi interface{} = m
		v.Set(reflect.ValueOf(oi))
		return nil
	}

	// check type of target. Currently on struct
	// TODO: support map?
	switch v.Kind() {
	case reflect.Struct:
		// ok
	default:
		return errors.New("json: cannot unmarshal object into Go value of type " + v.Type().String())
	}

	// Figure out field corresponding to key.
	var subv reflect.Value

	fields := cachedTypeFields(v.Type())
	for i := range fields {
		f := &fields[i]
		jsVal, ok := getMapValueForField(f, m)
		if !ok {
			continue
		}

		// TODO: no idea what this does
		subv = v
		for _, i := range f.index {
			if subv.Kind() == reflect.Ptr {
				if subv.IsNil() {
					// If a struct embeds a pointer to an unexported type,
					// it is not possible to set a newly allocated value
					// since the field is unexported.
					//
					// See https://golang.org/issue/21357
					if !subv.CanSet() {
						return fmt.Errorf("json: cannot set embedded pointer to unexported struct: %v", subv.Type().Elem())
					}
					subv.Set(reflect.New(subv.Type().Elem()))
				}
				subv = subv.Elem()
			}
			subv = subv.Field(i)
		}
		if err := jsonDecodeValueLax(subv, jsVal); err != nil {
			return err
		}
	}
	return nil
}

// TODO: should allow e.g. setting to int or string?
func jsonSetBoolLax(v reflect.Value, b bool) bool {
	switch v.Kind() {
	case reflect.Bool:
		v.Set(reflect.ValueOf(b))
		return true
	}
	return false
}

func jsonSetFloat64Lax(v reflect.Value, f float64) bool {
	switch v.Kind() {
	case reflect.Float64:
		v.Set(reflect.ValueOf(f))
		return true
	case reflect.Float32:
		v.Set(reflect.ValueOf(float32(f)))
		return true
	case reflect.Int:
		v.Set(reflect.ValueOf(int(f)))
		return true
		// TODO: more types
	}
	return false
}

func jsonSetStringLax(v reflect.Value, s string) bool {
	switch v.Kind() {
	case reflect.String:
		v.Set(reflect.ValueOf(s))
		return true
	}
	return false
}

func jsonSetArrayLax(v reflect.Value, a []interface{}) bool {
	panic("NYI")
}

func jsonSetObjectLax(v reflect.Value, o map[string]interface{}) bool {
	panic("NYI")
}

func jsonDecodeValueLax(v reflect.Value, jsVal interface{}) error {
	switch realJsVal := jsVal.(type) {
	case bool:
		jsonSetBoolLax(v, realJsVal)
	case float64:
		jsonSetFloat64Lax(v, realJsVal)
	case string:
		jsonSetStringLax(v, realJsVal)
	case []interface{}:
		jsonSetArrayLax(v, realJsVal)
	case map[string]interface{}:
		jsonSetObjectLax(v, realJsVal)
	default:
		panic(fmt.Sprintf("unsupported type %T of value %v", jsVal, jsVal))
	}
	return nil
}

func getMapValueForField(f *field, m map[string]interface{}) (interface{}, bool) {
	if v, ok := m[f.name]; ok {
		return v, true
	}
	// slow path, check every key
	for k, v := range m {
		if f.equalFold(f.nameBytes, []byte(k)) {
			return v, true
		}
	}
	return nil, false
}

// v can be *Foo if not nil or **Foo
func makeStructFromJSONMap3(v interface{}, m map[string]interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return errors.New("json: cannot unmarshal object into Go value of type " + rv.Type().String())
	}
	if rv.IsNil() {
		tname := rv.Type().String()
		return errors.New("json: cannot unmarshal object into Go nil value of type " + tname + " (either use " + tname + " to allocated struct or type *" + tname)
	}

	return jsonDecodeObject(rv, m)
}
