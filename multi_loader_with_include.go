package ravendb

import (
	"reflect"
)

// ILoaderWithInclude is NewMultiLoaderWithInclude

type MultiLoaderWithInclude struct {
	session  *DocumentSession
	includes []string
}

func NewMultiLoaderWithInclude(session *DocumentSession) *MultiLoaderWithInclude {
	return &MultiLoaderWithInclude{
		session: session,
	}
}

func (l *MultiLoaderWithInclude) Include(path string) *MultiLoaderWithInclude {
	l.includes = append(l.includes, path)
	return l
}

// results should be map[string]*struct
func (l *MultiLoaderWithInclude) LoadMulti(results interface{}, ids []string) error {
	if len(ids) == 0 {
		return newIllegalArgumentError("ids cannot be empty array")
	}
	if err := checkValidLoadMultiArg(results, "results"); err != nil {
		return err
	}

	return l.session.loadInternalMulti(results, ids, l.includes)
}

// TODO: needs a test
// TODO: better implementation
func (l *MultiLoaderWithInclude) Load(result interface{}, id string) error {
	if id == "" {
		return newIllegalArgumentError("id cannot be empty string")
	}
	// TODO: should allow map[string]interface{} as argument? (and therefore use checkValidLoadArg)
	if err := checkIsPtrPtrStruct(result, "result"); err != nil {
		return err
	}

	// create a map[string]typeof(result)
	rt := reflect.TypeOf(result)
	rt = rt.Elem() // it's now ptr-to-struct

	mapType := reflect.MapOf(stringType, rt)
	m := reflect.MakeMap(mapType)
	ids := []string{id}
	err := l.session.loadInternalMulti(m.Interface(), ids, l.includes)
	if err != nil {
		return err
	}
	key := reflect.ValueOf(id)
	res := m.MapIndex(key)
	if res.IsNil() {
		//return ErrNotFound
		return nil
	}
	return setInterfaceToValue(result, res.Interface())
}
