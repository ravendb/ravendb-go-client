package ravendb

import "reflect"

// TODO: there's only one implementation in MultiLoaderWithInclude so could simplify with:
// type ILoaderWithInclude = *MultiLoaderWithInclude
type ILoaderWithInclude interface {
	include(path string) ILoaderWithInclude
	load(clazz reflect.Type, id string) (interface{}, error)
	loadMulti(clazz reflect.Type, ids []string) (map[string]interface{}, error)
}
