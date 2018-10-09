package ravendb

import (
	"encoding/json"
	"reflect"
)

var _ ILazyOperation = &LazyLoadOperation{}

type LazyLoadOperation struct {
	_clazz         reflect.Type
	_session       *InMemoryDocumentSessionOperations
	_loadOperation *LoadOperation
	_ids           []string
	_includes      []string

	result        Object
	queryResult   *QueryResult
	requiresRetry bool
}

func NewLazyLoadOperation(clazz reflect.Type, session *InMemoryDocumentSessionOperations, loadOperation *LoadOperation) *LazyLoadOperation {
	return &LazyLoadOperation{
		_clazz:         clazz,
		_session:       session,
		_loadOperation: loadOperation,
	}
}

func (o *LazyLoadOperation) createRequest() *GetRequest {
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

func (o *LazyLoadOperation) byId(id string) *LazyLoadOperation {
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

func (o *LazyLoadOperation) getResult() Object {
	return o.result
}

func (o *LazyLoadOperation) setResult(result Object) {
	o.result = result
}

func (o *LazyLoadOperation) getQueryResult() *QueryResult {
	return o.queryResult
}

func (o *LazyLoadOperation) setQueryResult(queryResult *QueryResult) {
	o.queryResult = queryResult
}

func (o *LazyLoadOperation) isRequiresRetry() bool {
	return o.requiresRetry
}

func (o *LazyLoadOperation) setRequiresRetry(requiresRetry bool) {
	o.requiresRetry = requiresRetry
}

func (o *LazyLoadOperation) handleResponse(response *GetResponse) error {
	if response.isForceRetry {
		o.result = nil
		o.requiresRetry = true
		return nil
	}

	res := response.result
	if res == "" {
		o.handleResponse2(nil)
		return nil
	}
	var multiLoadResult *GetDocumentsResult
	err := json.Unmarshal([]byte(res), &multiLoadResult)
	if err != nil {
		return err
	}
	o.handleResponse2(multiLoadResult)
	return nil
}

func (o *LazyLoadOperation) handleResponse2(loadResult *GetDocumentsResult) {
	o._loadOperation.setResult(loadResult)

	if !o.requiresRetry {
		o.result = o._loadOperation.getDocuments(o._clazz)
	}
}
