package ravendb

// EvictItemsFromCacheBasedOnChanges is for evicting cache items
// based on database changes
type EvictItemsFromCacheBasedOnChanges struct {
	databaseName                string
	changes                     *DatabaseChanges
	documentsSubscriptionCloser CancelFunc
	indexesSubscriptionCloser   CancelFunc
	requestExecutor             *RequestExecutor
}

// NewEvictItemsFromCacheBasedOnChanges returns EvictItemsFromCacheBasedOnChanges
func NewEvictItemsFromCacheBasedOnChanges(store *DocumentStore, databaseName string) (*EvictItemsFromCacheBasedOnChanges, error) {
	res := &EvictItemsFromCacheBasedOnChanges{
		databaseName:    databaseName,
		changes:         store.Changes(databaseName),
		requestExecutor: store.GetRequestExecutor(databaseName),
	}

	docChange := func(documentChange *DocumentChange) {
		tp := documentChange.Type
		if tp == DocumentChangePut || tp == DocumentChangeDelete {
			cache := res.requestExecutor.Cache
			cache.generation.incrementAndGet()
		}
	}

	var err error
	res.documentsSubscriptionCloser, err = res.changes.ForAllDocuments(docChange)
	if err != nil {
		return nil, err
	}
	indexChange := func(indexChange *IndexChange) {
		tp := indexChange.Type
		if tp == IndexChangeBatchCompleted || tp == IndexChangeIndexRemoved {
			cache := res.requestExecutor.Cache
			cache.generation.incrementAndGet()
		}
	}
	res.indexesSubscriptionCloser, err = res.changes.ForAllIndexes(indexChange)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// Close closes EvictItemsFromCacheBasedOnChanges
func (e *EvictItemsFromCacheBasedOnChanges) Close() {
	e.documentsSubscriptionCloser()
	e.indexesSubscriptionCloser()
	e.changes.Close()
}
