package ravendb

import "reflect"

// TODO: there's only one implementation in MultiLoaderWithInclude so could simplify with:
// type ILoaderWithInclude = *MultiLoaderWithInclude
type ILoaderWithInclude interface {
	Include(path string) ILoaderWithInclude
	Load(clazz reflect.Type, id string) (interface{}, error)
	LoadMulti(clazz reflect.Type, ids []string) (map[string]interface{}, error)
}
