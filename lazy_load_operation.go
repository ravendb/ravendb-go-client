package ravendb

var _ ILazyOperation = &LazyLoadOperation{}

// LazyLoadOperation represents lazy load operation
type LazyLoadOperation struct {
	_session       *InMemoryDocumentSessionOperations
	_loadOperation *LoadOperation
	_ids           []string
	_includes      []string

	hasItems bool
	queryResult   *QueryResult
	//result interface{}
	requiresRetry bool
}

// NewLazyLoadOperation returns new LazyLoadOperation
func newLazyLoadOperation(session *InMemoryDocumentSessionOperations, loadOperation *LoadOperation) *LazyLoadOperation {
	return &LazyLoadOperation{
		_session:       session,
		_loadOperation: loadOperation,
	}
}

// TODO: should return an error
// needed for ILazyOperation
func (o *LazyLoadOperation) createRequest() *getRequest {
	var idsToCheckOnServer []string
	for _, id := range o._ids {
		if !o._session.IsLoadedOrDeleted(id) {
			idsToCheckOnServer = append(idsToCheckOnServer, id)
		}
	}
	queryBuilder := "?"

	if o._includes != nil {
		for _, include := range o._includes {
			queryBuilder += "&include="
			queryBuilder += include
		}
	}

	for _, id := range idsToCheckOnServer {
		queryBuilder += "&id="
		queryBuilder += urlUtilsEscapeDataString(id)
	}

	o.hasItems = len(idsToCheckOnServer) == 0

	if o.hasItems {
		// no need to hit the server
		return nil
	}

	getRequest := &getRequest{
		url:   "/docs",
		query: queryBuilder,
	}
	return getRequest
}

func (o *LazyLoadOperation) byID(id string) *LazyLoadOperation {
	if id == "" {
		return o
	}

	if o._ids == nil {
		o._ids = []string{id}
	}

	return o
}

func (o *LazyLoadOperation) byIds(ids []string) *LazyLoadOperation {
	o._ids = ids
	return o
}

func (o *LazyLoadOperation) withIncludes(includes []string) *LazyLoadOperation {
	o._includes = includes
	return o
}

// needed for ILazyOperation
func (o *LazyLoadOperation) getResult(result interface{}) error {
	if o.hasItems {
		return o._loadOperation.getDocuments(result)
	}

	if o.requiresRetry {
		return nil
	}

	err := o._loadOperation.getDocuments(result)
	// TODO: a better way to distinguish between a Load() and LoadMulti() operation
	if err != nil && len(o._ids) == 1 {
		err = o._loadOperation.getDocument(result)
	}
	return err
}

// needed for ILazyOperation
func (o *LazyLoadOperation) getQueryResult() *QueryResult {
	return o.queryResult
}

// needed for ILazyOperation
func (o *LazyLoadOperation) isRequiresRetry() bool {
	return o.requiresRetry
}

// needed for ILazyOperation
func (o *LazyLoadOperation) handleResponse(response *GetResponse) error {
	if response.IsForceRetry {
		o.requiresRetry = true
		return nil
	}

	res := response.Result
	if len(res) == 0 {
		return o.handleResponse2(nil)
	}
	var multiLoadResult *GetDocumentsResult
	err := jsonUnmarshal(res, &multiLoadResult)
	if err != nil {
		return err
	}
	return o.handleResponse2(multiLoadResult)
}

func (o *LazyLoadOperation) handleResponse2(loadResult *GetDocumentsResult) error {
	o._loadOperation.setResult(loadResult)
	// Note: rest delayed to getResult()
	return nil
}
