package ravendb

// Note: ILazySessionOperations is LazySessionOperations

// LazySessionOperations describes API for lazy operations
type LazySessionOperations struct {
	delegate *DocumentSession
}

func newLazySessionOperations(delegate *DocumentSession) *LazySessionOperations {
	return &LazySessionOperations{
		delegate: delegate,
	}
}

// Include adds a given object path to be included in results
func (o *LazySessionOperations) Include(path string) *LazyMultiLoaderWithInclude {
	return NewLazyMultiLoaderWithInclude(o.delegate).Include(path)
}

func (o *LazySessionOperations) LoadWithEval(id string, onEval func(), onEvalResult interface{}) (*Lazy, error) {
	if id == "" {
		return nil, newIllegalArgumentError("id cannot be empty string")
	}
	if o.delegate.IsLoaded(id) {
		fn := func(result interface{}) error {
			// TODO: test for this code path
			return o.delegate.Load(result, id)
		}
		return newLazy(fn), nil
	}

	session := o.delegate.InMemoryDocumentSessionOperations
	op := NewLoadOperation(session).byID(id)
	lazyLoadOperation := newLazyLoadOperation(session, op).byID(id)
	return o.delegate.addLazyOperation(lazyLoadOperation, onEval, onEvalResult), nil
}

func (o *LazySessionOperations) Load(id string) (*Lazy, error) {
	return o.LoadWithEval(id, nil, nil)
}

// LoadStartingWith returns Lazy object for lazily loading multiple value
// of a given type and matching args
// results should be map[string]*Struct
func (o *LazySessionOperations) LoadStartingWith(args *StartsWithArgs) *Lazy {
	session := o.delegate.InMemoryDocumentSessionOperations
	operation := NewLazyStartsWithOperation(args.StartsWith, args.Matches, args.Exclude, args.Start, args.PageSize, session, args.StartAfter)

	return o.delegate.addLazyOperation(operation, nil, nil)
}

// LoadMulti returns Lazy object for lazily loading multiple values
// of a given type and with given ids
func (o *LazySessionOperations) LoadMulti(ids []string) (*Lazy, error) {
	if len(ids) == 0 {
		return nil, newIllegalArgumentError("ids cannot be empty array")
	}
	return o.delegate.lazyLoadInternal(ids, nil, nil, nil), nil
}

func (o *LazySessionOperations) LoadMultiWithEval(ids []string, onEval func(), onEvalResult interface{}) (*Lazy, error) {
	if len(ids) == 0 {
		return nil, newIllegalArgumentError("ids cannot be empty array")
	}
	return o.delegate.lazyLoadInternal(ids, nil, onEval, onEvalResult), nil
}
