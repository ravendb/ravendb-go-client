package ravendb

import (
	"errors"
	"reflect"
)

// TODO: remove in API cleanup phase
type ILazySessionOperations = LazySessionOperations

type LazySessionOperations struct {
	delegate *DocumentSession
}

func NewLazySessionOperations(delegate *DocumentSession) *LazySessionOperations {
	return &LazySessionOperations{
		delegate: delegate,
	}
}

func (o *LazySessionOperations) Include(path string) *ILazyLoaderWithInclude {
	return NewLazyMultiLoaderWithInclude(o.delegate).Include(path)
}

// Lazy<TResult>
func (o *LazySessionOperations) Load(clazz reflect.Type, id string, onEval func(interface{})) *Lazy {
	if o.delegate.IsLoaded(id) {
		fn := func() (interface{}, error) {
			//return o.delegate.LoadOld(clazz, id)
			return nil, errors.New("NYI")
		}
		return NewLazy(fn)
	}

	session := o.delegate.InMemoryDocumentSessionOperations
	op := NewLoadOperation(session)
	op = op.byId(id)
	lazyLoadOperation := NewLazyLoadOperation(clazz, session, op).byId(id)
	return o.delegate.addLazyOperation(clazz, lazyLoadOperation, onEval)
}

func (o *LazySessionOperations) LoadStartingWith(clazz reflect.Type, args *StartsWithArgs) *Lazy {
	session := o.delegate.InMemoryDocumentSessionOperations
	operation := NewLazyStartsWithOperation(clazz, args.StartsWith, args.Matches, args.Exclude, args.Start, args.PageSize, session, args.StartAfter)

	t := reflect.MapOf(stringType, clazz)
	return o.delegate.addLazyOperation(t, operation, nil)
}

func (o *LazySessionOperations) LoadMulti(clazz reflect.Type, ids []string, onEval func(interface{})) *Lazy {
	return o.delegate.lazyLoadInternal(clazz, ids, nil, onEval)
}
