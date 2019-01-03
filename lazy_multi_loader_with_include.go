package ravendb

import (
	"fmt"
	"reflect"
)

// Note: ILazyLoaderWithInclude is LazyMultiLoaderWithInclude

// LazyMultiLoaderWithInclude is for lazily loading one or more objects with includes
type LazyMultiLoaderWithInclude struct {
	_session  *DocumentSession
	_includes []string
}

// NewLazyMultiLoaderWithInclude creates a lazy multi loader with includes
func NewLazyMultiLoaderWithInclude(session *DocumentSession) *LazyMultiLoaderWithInclude {
	return &LazyMultiLoaderWithInclude{
		_session: session,
	}
}

// Include adds ids of objects to add in a request
func (l *LazyMultiLoaderWithInclude) Include(path string) *LazyMultiLoaderWithInclude {
	l._includes = append(l._includes, path)
	return l
}

// LoadMulti lazily loads multiple values of a given type with given ids
// TODO: not covered by tests at all
func (l *LazyMultiLoaderWithInclude) LoadMulti(results interface{}, ids []string) *Lazy {
	return l._session.lazyLoadInternal(results, ids, l._includes, nil)
}

// m is a single-element map[string]*struct
// returns single map value
func getOneMapValue(results interface{}) (interface{}, error) {
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

// Load lazy loads a value with a given id into result
func (l *LazyMultiLoaderWithInclude) Load(result interface{}, id string) *Lazy {
	ids := []string{id}
	// result should be **Foo, make map[string]*Foo

	tp := reflect.TypeOf(result)
	if tp.Kind() != reflect.Ptr && tp.Elem().Kind() != reflect.Ptr {
		// TODO: return as an error
		panicIf(true, "expected result to be **Foo, is %T", result)
	}

	resultType := reflect.MapOf(stringType, tp.Elem())
	results := reflect.MakeMap(resultType).Interface()

	lazy := l._session.lazyLoadInternal(results, ids, l._includes, nil)
	valueFactory := func(result2 interface{}) error {
		panicIf(reflect.TypeOf(result) != reflect.TypeOf(result2), "LazyMultiLoaderWithInclude.Load(): expected values of same, type, got: result=%T, result2=%T\n", result, result2)
		err := lazy.GetValue()
		if err != nil {
			return err
		}

		m := reflect.ValueOf(results)
		if m.Len() == 0 {
			return nil
		}
		panicIf(m.Len() != 1, "expected m to have size of 1, got %d", m.Len())

		key := reflect.ValueOf(id)
		res := m.MapIndex(key)
		if res.IsNil() {
			return ErrNotFound
		}
		setInterfaceToValue(result, res.Interface())
		return nil
	}
	return NewLazy(result, valueFactory)
}
