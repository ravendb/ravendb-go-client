package ravendb

import (
	"errors"
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
func (o *LazySessionOperations) Include(path string) *ILazyLoaderWithInclude {
	return NewLazyMultiLoaderWithInclude(o.delegate).Include(path)
}

// Load returns Lazy object for lazily loading a value of a given type and id
// from the database
func (o *LazySessionOperations) Load(clazz reflect.Type, id string, onEval func(interface{})) *Lazy {
	if o.delegate.IsLoaded(id) {
		fn := func() (interface{}, error) {
			//return o.delegate.LoadOld(clazz, id)
			panic("NYI")
			return nil, errors.New("NYI")
		}
		return NewLazy(fn)
	}

	session := o.delegate.InMemoryDocumentSessionOperations
	op := NewLoadOperation(session).byID(id)
	lazyLoadOperation := NewLazyLoadOperation(clazz, session, op).byID(id)
	return o.delegate.addLazyOperation(clazz, lazyLoadOperation, onEval)
}

// LoadStartingWith returns Lazy object for lazily loading mutliple value
// of a given type and matching args
func (o *LazySessionOperations) LoadStartingWith(clazz reflect.Type, args *StartsWithArgs) *Lazy {
	session := o.delegate.InMemoryDocumentSessionOperations
	operation := NewLazyStartsWithOperation(clazz, args.StartsWith, args.Matches, args.Exclude, args.Start, args.PageSize, session, args.StartAfter)

	t := reflect.MapOf(stringType, clazz)
	return o.delegate.addLazyOperation(t, operation, nil)
}

// LoadMulti returns Lazy object for lazily loading multiple values
// of a given type and with given ids
func (o *LazySessionOperations) LoadMulti(clazz reflect.Type, ids []string, onEval func(interface{})) *Lazy {
	return o.delegate.lazyLoadInternal(clazz, ids, nil, onEval)
}
