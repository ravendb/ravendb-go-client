package ravendb

import (
	"reflect"
	"strings"
)

// tryGetIDFromInstance returns value of ID field on struct if it's of type
// string. Returns empty string if there's no ID field or it's not string
func tryGetIDFromInstance(entity interface{}) (string, bool) {
	rv := reflect.ValueOf(entity)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		// TODO: maybe panic?
		return "", false
	}
	structType := rv.Type()
	nFields := rv.NumField()
	for i := 0; i < nFields; i++ {
		structField := structType.Field(i)
		name := structField.Name
		if len(name) != 2 {
			// perf micro-optimization
			continue
		}
		// TODO: maybe only allow ID?
		if !strings.EqualFold(name, "id") {
			continue
		}
		// don't read unexported (starting with lowercase) fields
		// TODO: there's probably a better way
		if name[0] == 'i' {
			continue
		}
		// TODO: allow int type for id?
		if structField.Type.Kind() != reflect.String {
			continue
		}
		return rv.Field(i).String(), true
	}
	return "", false
}

// trySetIDOnEnity tries to set value of ID field on struct to id
// returns false if entity has no ID field or if it's not string
func trySetIDOnEntity(entity interface{}, id string) bool {
	rv := reflect.ValueOf(entity)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		// TODO: maybe panic?
		return false
	}
	structType := rv.Type()
	nFields := rv.NumField()
	for i := 0; i < nFields; i++ {
		structField := structType.Field(i)
		name := structField.Name
		if len(name) != 2 {
			// perf micro-optimization
			continue
		}
		// TODO: maybe only allow ID?
		if !strings.EqualFold(name, "id") {
			continue
		}
		// TODO: allow int type for id?
		if structField.Type.Kind() != reflect.String {
			continue
		}
		field := rv.Field(i)
		if !field.CanSet() {
			return false
		}
		rv.Field(i).SetString(id)
		return true
	}
	return false
}
