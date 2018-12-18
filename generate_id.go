package ravendb

import (
	"reflect"
)

func tryGetIDFromInstance(entity interface{}) (string, bool) {
	return TryGetIDFromInstance(entity)
}

// TryGetIDFromInstance returns value of ID field on struct if it's of type
// string. Returns empty string if there's no ID field or it's not string
func TryGetIDFromInstance(entity interface{}) (string, bool) {
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
		if name != "ID" {
			continue
		}
		if structField.Type.Kind() != reflect.String {
			continue
		}
		return rv.Field(i).String(), true
	}
	return "", false
}

// trySetIDOnEnity tries to set value of ID field on struct to id
// returns false if entity has no ID field or if it's not string
func TrySetIDOnEntity(entity interface{}, id string) bool {
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
		if name != "ID" {
			continue
		}
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
