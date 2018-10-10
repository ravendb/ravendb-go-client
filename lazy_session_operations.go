package ravendb

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

/*
    public ILazyLoaderWithInclude include(String path) {
        return new LazyMultiLoaderWithInclude(delegate).include(path);
    }

    public <TResult> Lazy<TResult> load(Class<TResult> clazz, String id) {
        return load(clazz, id, null);
    }

    public <TResult> Lazy<TResult> load(Class<TResult> clazz, String id, Consumer<TResult> onEval) {
        if (delegate.isLoaded(id)) {
            return new Lazy<>(() -> delegate.load(clazz, id));
        }

        LazyLoadOperation lazyLoadOperation = new LazyLoadOperation(clazz, delegate, new LoadOperation(delegate).byId(id)).byId(id);
        return delegate.addLazyOperation(clazz, lazyLoadOperation, onEval);
    }

    public <TResult> Lazy<Map<String, TResult>> loadStartingWith(Class<TResult> clazz, String idPrefix) {
        return loadStartingWith(clazz, idPrefix, null, 0, 25, null, null);
    }

    public <TResult> Lazy<Map<String, TResult>> loadStartingWith(Class<TResult> clazz, String idPrefix, String matches) {
        return loadStartingWith(clazz, idPrefix, matches, 0, 25, null, null);
    }

    public <TResult> Lazy<Map<String, TResult>> loadStartingWith(Class<TResult> clazz, String idPrefix, String matches, int start) {
        return loadStartingWith(clazz, idPrefix, matches, start, 25, null, null);
    }

    public <TResult> Lazy<Map<String, TResult>> loadStartingWith(Class<TResult> clazz, String idPrefix, String matches, int start, int pageSize) {
        return loadStartingWith(clazz, idPrefix, matches, start, pageSize, null, null);
    }

    public <TResult> Lazy<Map<String, TResult>> loadStartingWith(Class<TResult> clazz, String idPrefix, String matches, int start, int pageSize, String exclude) {
        return loadStartingWith(clazz, idPrefix, matches, start, pageSize, exclude, null);
    }

    public <TResult> Lazy<Map<String, TResult>> loadStartingWith(Class<TResult> clazz, String idPrefix, String matches, int start, int pageSize, String exclude, String startAfter) {
        LazyStartsWithOperation operation = new LazyStartsWithOperation<>(clazz, idPrefix, matches, exclude, start, pageSize, delegate, startAfter);

        return delegate.addLazyOperation((Class<Map<String, TResult>>)(Class<?>)Map.class, operation, null);
    }

    public <TResult> Lazy<Map<String, TResult>> load(Class<TResult> clazz, Collection<String> ids) {
        return load(clazz, ids, null);
    }

    public <TResult> Lazy<Map<String, TResult>> load(Class<TResult> clazz, Collection<String> ids, Consumer<Map<String, TResult>> onEval) {
        return delegate.lazyLoadInternal(clazz, ids.toArray(new String[0]), new String[0], onEval);
    }

    //TBD expr ILazyLoaderWithInclude<T> ILazySessionOperations.Include<T>(Expression<Func<T, string>> path)
    //TBD expr ILazyLoaderWithInclude<T> ILazySessionOperations.Include<T>(Expression<Func<T, IEnumerable<string>>> path)
}
*/
