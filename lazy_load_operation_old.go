package ravendb

import (
	"reflect"
)

var _ ILazyOperation = &LazyLoadOperationOld{}

// LazyLoadOperationOld represents lazy load operation
type LazyLoadOperationOld struct {
	_clazz         reflect.Type
	_session       *InMemoryDocumentSessionOperations
	_loadOperation *LoadOperation
	_ids           []string
	_includes      []string

	result        interface{}
	queryResult   *QueryResult
	requiresRetry bool
}

// NewLazyLoadOperation returns new LazyLoadOperationOld
func NewLazyLoadOperationOld(clazz reflect.Type, session *InMemoryDocumentSessionOperations, loadOperation *LoadOperation) *LazyLoadOperationOld {
	return &LazyLoadOperationOld{
		_clazz:         clazz,
		_session:       session,
		_loadOperation: loadOperation,
	}
}

func (o *LazyLoadOperationOld) createRequest() *GetRequest {
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
		queryBuilder += UrlUtils_escapeDataString(id)
	}

	hasItems := len(idsToCheckOnServer) > 0

	if !hasItems {
		// no need to hit the server
		o.result = o._loadOperation.getDocuments(o._clazz)
		return nil
	}

	getRequest := &GetRequest{
		url:   "/docs",
		query: queryBuilder,
	}
	return getRequest
}

func (o *LazyLoadOperationOld) byID(id string) *LazyLoadOperationOld {
	if id == "" {
		return o
	}

	if o._ids == nil {
		o._ids = []string{id}
	}

	return o
}

func (o *LazyLoadOperationOld) byIds(ids []string) *LazyLoadOperationOld {
	o._ids = ids

	return o
}

func (o *LazyLoadOperationOld) withIncludes(includes []string) *LazyLoadOperationOld {
	o._includes = includes
	return o
}

// needed for ILazyOperation
func (o *LazyLoadOperationOld) getResult() interface{} {
	return o.result
}

// needed for ILazyOperation
func (o *LazyLoadOperationOld) getQueryResult() *QueryResult {
	return o.queryResult
}

// needed for ILazyOperation
func (o *LazyLoadOperationOld) isRequiresRetry() bool {
	return o.requiresRetry
}

func (o *LazyLoadOperationOld) handleResponse(response *GetResponse) error {
	if response.isForceRetry {
		o.result = nil
		o.requiresRetry = true
		return nil
	}

	res := response.result
	if len(res) == 0 {
		o.handleResponse2(nil)
		return nil
	}
	var multiLoadResult *GetDocumentsResult
	err := jsonUnmarshal(res, &multiLoadResult)
	if err != nil {
		return err
	}
	return o.handleResponse2(multiLoadResult)
}

func (o *LazyLoadOperationOld) handleResponse2(loadResult *GetDocumentsResult) error {
	o._loadOperation.setResult(loadResult)

	var err error
	if !o.requiresRetry {
		o.result, err = o._loadOperation.getDocumentsOld(o._clazz)
	}
	return err
}
