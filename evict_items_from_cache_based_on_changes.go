package ravendb

import "io"

// EvictItemsFromCacheBasedOnChanges is for evicting cache items
// based on database changes
type EvictItemsFromCacheBasedOnChanges struct {
	_databaseName          string
	_changes               *DatabaseChanges
	_documentsSubscription io.Closer
	_indexesSubscription   io.Closer
	_requestExecutor       *RequestExecutor
}

func NewEvictItemsFromCacheBasedOnChanges(store *DocumentStore, databaseName string) *EvictItemsFromCacheBasedOnChanges {
	res := &EvictItemsFromCacheBasedOnChanges{
		_databaseName:    databaseName,
		_changes:         store.ChangesWithDatabaseName(databaseName),
		_requestExecutor: store.GetRequestExecutor(databaseName),
	}
	docSub, err := res._changes.ForAllDocuments()
	must(err) // TODO: return an error?
	res._documentsSubscription = docSub.Subscribe(res)
	indexSub, err := res._changes.ForAllIndexes()
	must(err) // TODO: return an error?
	res._indexesSubscription = indexSub.Subscribe(res)
	return res
}

func (e *EvictItemsFromCacheBasedOnChanges) OnNext(value interface{}) {
	if documentChange, ok := value.(*DocumentChange); ok {
		tp := documentChange.Type
		if tp == DocumentChangePut || tp == DocumentChangeDelete {
			cache := e._requestExecutor.Cache
			cache.generation.incrementAndGet()
		}
	} else if indexChange, ok := value.(*IndexChange); ok {
		tp := indexChange.Type
		if tp == IndexChangeBatchCompleted || tp == IndexChangeIndexRemoved {
			cache := e._requestExecutor.Cache
			cache.generation.incrementAndGet()
		}
	}
}

func (e *EvictItemsFromCacheBasedOnChanges) OnError(err error) {
	// empty
}

func (e *EvictItemsFromCacheBasedOnChanges) OnCompleted() {
	// empty
}

func (e *EvictItemsFromCacheBasedOnChanges) Close() {
	changesScope := e._changes
	defer changesScope.Close()

	e._documentsSubscription.Close()
	e._indexesSubscription.Close()
}
