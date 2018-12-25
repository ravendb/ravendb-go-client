package ravendb

import "reflect"

// TODO: verify this is the same as Java
func isPrimitiveOrWrapper(t reflect.Type) bool {
	kind := t.Kind()

	/*
		Uintptr
		Complex64
		Complex128
		Array
		Chan
		Func
		Interface
		Map
		Ptr
		Slice
		Struct
		UnsafePointer
	*/
	switch kind {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64, reflect.String:
		return true
	}
	return false
}

// Go doesn't have enums
func typeIsEnum(t reflect.Type) bool {
	return false
}
