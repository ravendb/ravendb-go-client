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

func NewEvictItemsFromCacheBasedOnChanges(store *DocumentStore, databaseName string) (*EvictItemsFromCacheBasedOnChanges, error) {
	res := &EvictItemsFromCacheBasedOnChanges{
		_databaseName:    databaseName,
		_changes:         store.ChangesWithDatabaseName(databaseName),
		_requestExecutor: store.GetRequestExecutor(databaseName),
	}
	docSub, err := res._changes.ForAllDocuments()
	if err != nil {
		return nil, err
	}
	res._documentsSubscription = docSub.Subscribe(res)
	indexSub, err := res._changes.ForAllIndexes()
	if err != nil {
		return nil, err
	}
	res._indexesSubscription = indexSub.Subscribe(res)
	return res, nil
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
