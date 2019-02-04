package ravendb

import (
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
func (l *LazyMultiLoaderWithInclude) LoadMulti(results interface{}, ids []string) (*Lazy, error) {
	if len(ids) == 0 {
		return nil, newIllegalArgumentError("ids cannot be empty array")
	}
	if err := checkValidLoadMultiArg(results, "results"); err != nil {
		return nil, err
	}

	return l._session.lazyLoadInternal(results, ids, l._includes, nil), nil
}

// Load lazy loads a value with a given id into result
func (l *LazyMultiLoaderWithInclude) Load(result interface{}, id string) (*Lazy, error) {
	if id == "" {
		return nil, newIllegalArgumentError("id cannot be empty string")
	}
	// TODO: should allow map[string]interface{} as argument? (and therefore use checkValidLoadArg)
	if err := checkIsPtrPtrStruct(result, "result"); err != nil {
		return nil, err
	}

	ids := []string{id}
	// result should be **Foo, make map[string]*Foo

	tp := reflect.TypeOf(result)
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
	return NewLazy(result, valueFactory), nil
}
