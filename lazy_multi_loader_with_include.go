package ravendb

import (
	"errors"
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
func (l *LazyMultiLoaderWithInclude) LoadMulti(clazz reflect.Type, ids []string) *Lazy {
	return l._session.lazyLoadInternal(clazz, ids, l._includes, nil)
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

// Load lazily loads a value of a given type with a given id
func (l *LazyMultiLoaderWithInclude) Load(clazz reflect.Type, id string) *Lazy {
	ids := []string{id}
	results := l._session.lazyLoadInternal(clazz, ids, l._includes, nil)
	valueFactory := func() (interface{}, error) {
		// resultMap is map[string]clazz
		resultMap, err := results.GetValue()
		if err != nil {
			return nil, err
		}
		if m, ok := resultMap.(map[string]interface{}); ok {
			if len(m) == 0 {
				return nil, nil
			}
			panicIf(len(m) != 1, "expected resultMap to only have one element, has %d", len(m))
			for _, v := range m {
				return v, nil
			}
			return nil, errors.New("Impossible!")
		}

		//fmt.Printf("resultMap: %v, type=%T\n", resultMap, resultMap)
		// TODO: not sure if this is needed
		v, err := getOneMapValue(resultMap)
		fmt.Printf("LazyMultiLoaderWithInclude.Load(): v=%v, type=%T\n", v, v)
		must(err)
		return v, nil
	}
	return NewLazy(valueFactory)
}
