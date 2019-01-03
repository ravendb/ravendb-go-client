package ravendb

import (
	"reflect"
)

// Note: ILazySessionOperations is LazySessionOperations

// LazySessionOperations describes API for lazy operations
type LazySessionOperations struct {
	delegate *DocumentSession
}

// NewLazySessionOperations returns new LazySessionOperations
func NewLazySessionOperations(delegate *DocumentSession) *LazySessionOperations {
	return &LazySessionOperations{
		delegate: delegate,
	}
}

// Include adds a given object path to be included in results
func (o *LazySessionOperations) Include(path string) *LazyMultiLoaderWithInclude {
	return NewLazyMultiLoaderWithInclude(o.delegate).Include(path)
}

func (o *LazySessionOperations) Load(result interface{}, id string, onEval func(interface{})) *Lazy {
	if o.delegate.IsLoaded(id) {
		fn := func(result interface{}) error {
			// TODO: test for this code path
			return o.delegate.Load(result, id)
		}
		return NewLazy2(result, fn)
	}

	session := o.delegate.InMemoryDocumentSessionOperations
	op := NewLoadOperation(session).byID(id)
	lazyLoadOperation := NewLazyLoadOperation(result, session, op).byID(id)
	return o.delegate.addLazyOperation(result, lazyLoadOperation, onEval)
}

// LoadStartingWith returns Lazy object for lazily loading multiple value
// of a given type and matching args
func (o *LazySessionOperations) LoadStartingWithOld(clazz reflect.Type, args *StartsWithArgs) *Lazy {
	session := o.delegate.InMemoryDocumentSessionOperations
	operation := NewLazyStartsWithOperation(clazz, args.StartsWith, args.Matches, args.Exclude, args.Start, args.PageSize, session, args.StartAfter)

	t := reflect.MapOf(stringType, clazz)
	return o.delegate.addLazyOperationOld(t, operation, nil)
}

// LoadMulti returns Lazy object for lazily loading multiple values
// of a given type and with given ids
func (o *LazySessionOperations) LoadMulti(results interface{}, ids []string, onEval func(interface{})) *Lazy {
	return o.delegate.lazyLoadInternal(results, ids, nil, onEval)
}
