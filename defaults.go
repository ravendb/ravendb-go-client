package ravendb

import "reflect"

// Go port of com.google.common.base.Defaults to make porting Java easier

func Defaults_defaultValue(clazz reflect.Type) interface{} {
	panicIf(true, "NYI")
	return nil
}
