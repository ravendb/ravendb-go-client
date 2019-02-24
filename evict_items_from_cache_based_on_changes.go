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

	var chDocChanges chan *DocumentChange
	var err error
	chDocChanges, res.documentsSubscriptionCloser, err = res.changes.ForAllDocuments()
	if err != nil {
		return nil, err
	}

	go func() {
		for documentChange := range chDocChanges {
			tp := documentChange.Type
			if tp == DocumentChangePut || tp == DocumentChangeDelete {
				cache := res.requestExecutor.Cache
				cache.generation.incrementAndGet()
			}
		}
	}()

	var chIdxChanges chan *IndexChange
	chIdxChanges, res.indexesSubscriptionCloser, err = res.changes.ForAllIndexes()
	if err != nil {
		res.Close()
		return nil, err
	}

	// TODO: maybe combine the 2 goroutines?
	go func() {
		for indexChange := range chIdxChanges {
			tp := indexChange.Type
			if tp == IndexChangeBatchCompleted || tp == IndexChangeIndexRemoved {
				cache := res.requestExecutor.Cache
				cache.generation.incrementAndGet()
			}
		}
	}()
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
