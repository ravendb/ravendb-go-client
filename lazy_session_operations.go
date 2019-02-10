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

func (o *LazySessionOperations) Load(result interface{}, id string, onEval func(interface{})) (*Lazy, error) {
	if id == "" {
		return nil, newIllegalArgumentError("id cannot be empty string")
	}
	// TODO: should allow map[string]interface{} as argument? (and therefore use checkValidLoadArg)
	if err := checkIsPtrPtrStruct(result, "result"); err != nil {
		return nil, err
	}
	if o.delegate.IsLoaded(id) {
		fn := func(result interface{}) error {
			// TODO: test for this code path
			return o.delegate.Load(result, id)
		}
		return newLazy(result, fn, nil), nil
	}

	session := o.delegate.InMemoryDocumentSessionOperations
	op := NewLoadOperation(session).byID(id)
	lazyLoadOperation := NewLazyLoadOperation(result, session, op).byID(id)
	return o.delegate.addLazyOperation(result, lazyLoadOperation, onEval), nil
}

// LoadStartingWith returns Lazy object for lazily loading multiple value
// of a given type and matching args
// results should be map[string]*Struct
func (o *LazySessionOperations) LoadStartingWith(results interface{}, args *StartsWithArgs) *Lazy {
	session := o.delegate.InMemoryDocumentSessionOperations
	operation := NewLazyStartsWithOperation(results, args.StartsWith, args.Matches, args.Exclude, args.Start, args.PageSize, session, args.StartAfter)

	return o.delegate.addLazyOperation(results, operation, nil)
}

// LoadMulti returns Lazy object for lazily loading multiple values
// of a given type and with given ids
func (o *LazySessionOperations) LoadMulti(results interface{}, ids []string, onEval func(interface{})) (*Lazy, error) {
	if len(ids) == 0 {
		return nil, newIllegalArgumentError("ids cannot be empty array")
	}
	if err := checkValidLoadMultiArg(results, "results"); err != nil {
		return nil, err
	}

	return o.delegate.lazyLoadInternal(results, ids, nil, onEval), nil
}
