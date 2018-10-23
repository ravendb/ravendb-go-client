package ravendb

import "reflect"

type LazyMultiLoaderWithInclude struct {
	_session  *DocumentSession
	_includes []string
}

func NewLazyMultiLoaderWithInclude(session *DocumentSession) *LazyMultiLoaderWithInclude {
	return &LazyMultiLoaderWithInclude{
		_session: session,
	}
}

func (l *LazyMultiLoaderWithInclude) Include(path string) *ILazyLoaderWithInclude {
	l._includes = append(l._includes, path)
	return l
}

// Lazy<Map<String, T>>
func (l *LazyMultiLoaderWithInclude) Load(clazz reflect.Type, ids []string) *Lazy {
	panic("NYI")
	//return l._session.lazyLoadInternal(clazz, ids, l._includes, nil)
	return nil
}

/*
   public <TResult> Lazy<TResult> load(Class<TResult> clazz, String id) {
       Lazy<Map<String, TResult>> results = _session.lazyLoadInternal(clazz, new String[]{id}, _includes.toArray(new String[0]), null);
       return new Lazy(() -> results.getValue().values().iterator().next());
   }
*/
