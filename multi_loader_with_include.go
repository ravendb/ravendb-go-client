package ravendb

import "reflect"

var _ ILoaderWithInclude = &MultiLoaderWithInclude{}

type MultiLoaderWithInclude struct {
	_session  *IDocumentSessionImpl
	_includes []string
}

func NewMultiLoaderWithInclude(session *IDocumentSessionImpl) *MultiLoaderWithInclude {
	return &MultiLoaderWithInclude{
		_session: session,
	}
}

func (l *MultiLoaderWithInclude) Include(path string) ILoaderWithInclude {
	l._includes = append(l._includes, path)
	return l
}

func (l *MultiLoaderWithInclude) LoadMulti(clazz reflect.Type, ids []string) (map[string]interface{}, error) {
	return l._session.LoadInternalMultiOld(clazz, ids, l._includes)
}

func (l *MultiLoaderWithInclude) Load(clazz reflect.Type, id string) (interface{}, error) {
	stringObjectMap, err := l._session.LoadInternalMultiOld(clazz, []string{id}, l._includes)
	if err != nil {
		return nil, err
	}
	for _, v := range stringObjectMap {
		if v != nil {
			return v, nil
		}
	}
	return nil, nil
}
