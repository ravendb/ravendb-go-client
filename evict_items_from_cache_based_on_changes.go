package ravendb

// evictItemsFromCacheBasedOnChanges is for evicting cache items
// based on database changes
type evictItemsFromCacheBasedOnChanges struct {
	databaseName                string
	changes                     *DatabaseChanges
	documentsSubscriptionCloser CancelFunc
	indexesSubscriptionCloser   CancelFunc
	requestExecutor             *RequestExecutor
}

// newEvictItemsFromCacheBasedOnChanges returns EvictItemsFromCacheBasedOnChanges
func newEvictItemsFromCacheBasedOnChanges(store *DocumentStore, databaseName string) (*evictItemsFromCacheBasedOnChanges, error) {
	res := &evictItemsFromCacheBasedOnChanges{
		databaseName:    databaseName,
		changes:         store.Changes(databaseName),
		requestExecutor: store.GetRequestExecutor(databaseName),
	}

	cbDocChange := func(documentChange *DocumentChange) {
		tp := documentChange.Type
		if tp == DocumentChangePut || tp == DocumentChangeDelete {
			cache := res.requestExecutor.Cache
			cache.incGeneration()
		}
	}

	var err error
	res.documentsSubscriptionCloser, err = res.changes.ForAllDocuments(cbDocChange)
	if err != nil {
		return nil, err
	}

	cbIndexChange := func(indexChange *IndexChange) {
		tp := indexChange.Type
		if tp == IndexChangeBatchCompleted || tp == IndexChangeIndexRemoved {
			cache := res.requestExecutor.Cache
			cache.incGeneration()
		}
	}

	res.indexesSubscriptionCloser, err = res.changes.ForAllIndexes(cbIndexChange)
	if err != nil {
		res.Close()
		return nil, err
	}

	return res, nil
}

// Close closes EvictItemsFromCacheBasedOnChanges
func (e *evictItemsFromCacheBasedOnChanges) Close() {
	if e.documentsSubscriptionCloser != nil {
		e.documentsSubscriptionCloser()
	}
	if e.indexesSubscriptionCloser != nil {
		e.indexesSubscriptionCloser()
	}
	e.changes.Close()
}
