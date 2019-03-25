package ravendb

// Note: ILazyLoaderWithInclude is LazyMultiLoaderWithInclude

// LazyMultiLoaderWithInclude is for lazily loading one or more objects with includes
type LazyMultiLoaderWithInclude struct {
	session  *DocumentSession
	includes []string
}

// NewLazyMultiLoaderWithInclude creates a lazy multi loader with includes
func NewLazyMultiLoaderWithInclude(session *DocumentSession) *LazyMultiLoaderWithInclude {
	return &LazyMultiLoaderWithInclude{
		session: session,
	}
}

// Include adds ids of objects to add in a request
func (l *LazyMultiLoaderWithInclude) Include(path string) *LazyMultiLoaderWithInclude {
	l.includes = append(l.includes, path)
	return l
}

// LoadMulti lazily loads multiple values of a given type with given ids
func (l *LazyMultiLoaderWithInclude) LoadMulti(ids []string) (*Lazy, error) {
	if len(ids) == 0 {
		return nil, newIllegalArgumentError("ids cannot be empty array")
	}
	return l.session.lazyLoadInternal(ids, l.includes, nil, nil), nil
}

// Load lazy loads a value with a given id into result
func (l *LazyMultiLoaderWithInclude) Load(id string) (*Lazy, error) {
	if id == "" {
		return nil, newIllegalArgumentError("id cannot be empty string")
	}
	ids := []string{id}
	// result should be **Foo, make map[string]*Foo

	lazy := l.session.lazyLoadInternal(ids, l.includes, nil, nil)
	valueFactory := func(result interface{}) error {
		return lazy.GetValue(result)
	}
	return newLazy(valueFactory), nil
}
