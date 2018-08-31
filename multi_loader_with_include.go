package ravendb

import (
	"fmt"
	"reflect"
)

type MultiLoaderWithInclude struct {
	_session  *DocumentSession
	_includes []string
}

func NewMultiLoaderWithInclude(session *DocumentSession) *MultiLoaderWithInclude {
	return &MultiLoaderWithInclude{
		_session: session,
	}
}

func (l *MultiLoaderWithInclude) Include(path string) *MultiLoaderWithInclude {
	l._includes = append(l._includes, path)
	return l
}

// results should be map[string]*struct
func (l *MultiLoaderWithInclude) LoadMulti(results interface{}, ids []string) error {
	return l._session.loadInternalMulti(results, ids, l._includes)
}

// TODO: needs a test
// TODO: better implementation
func (l *MultiLoaderWithInclude) Load(result interface{}, id string) error {
	// create a map[string]typeof(result)
	rt := reflect.TypeOf(result)
	if rt.Kind() != reflect.Ptr || rt.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("type of result should be pointer-to-struct but is %T", result)
	}

	mapType := reflect.MapOf(stringType, rt)
	m := reflect.MakeMap(mapType)
	ids := []string{id}
	err := l._session.loadInternalMulti(m.Interface(), ids, l._includes)
	if err != nil {
		return err
	}
	key := reflect.ValueOf(id)
	res := m.MapIndex(key)
	if res.IsNil() {
		return ErrNotFound
	}
	setInterfaceToValue(result, res.Interface())
	return nil
}
