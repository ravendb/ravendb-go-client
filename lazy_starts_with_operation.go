package ravendb

import (
	"fmt"
	"reflect"
	"strings"
)

var _ ILazyOperation = &LazyStartsWithOperation{}

type LazyStartsWithOperation struct {
	_clazz             reflect.Type
	_idPrefix          string
	_matches           string
	_exclude           string
	_start             int
	_pageSize          int
	_sessionOperations *InMemoryDocumentSessionOperations
	_startAfter        string

	result        interface{}
	queryResult   *QueryResult
	requiresRetry bool
}

// TODO: convert to use StartsWithArgs
func NewLazyStartsWithOperation(clazz reflect.Type, idPrefix string, matches string, exclude string, start int, pageSize int, sessionOperations *InMemoryDocumentSessionOperations, startAfter string) *LazyStartsWithOperation {
	return &LazyStartsWithOperation{
		_clazz:             clazz,
		_idPrefix:          idPrefix,
		_matches:           matches,
		_exclude:           exclude,
		_start:             start,
		_pageSize:          pageSize,
		_sessionOperations: sessionOperations,
		_startAfter:        startAfter,
	}
}

func (o *LazyStartsWithOperation) createRequest() *GetRequest {
	pageSize := o._pageSize
	if pageSize == 0 {
		pageSize = 25
	}
	q := fmt.Sprintf("?startsWith=%s&matches=%s&exclude=%s&start=%d&pageSize=%d&startAfter=%s",
		UrlUtils_escapeDataString(o._idPrefix),
		UrlUtils_escapeDataString(o._matches),
		UrlUtils_escapeDataString(o._exclude),
		o._start,
		pageSize,
		o._startAfter)

	request := &GetRequest{
		url:   "/docs",
		query: q,
	}

	return request
}

func (o *LazyStartsWithOperation) getResult() interface{} {
	return o.result
}

func (o *LazyStartsWithOperation) setResult(result interface{}) {
	o.result = result
}

func (o *LazyStartsWithOperation) getQueryResult() *QueryResult {
	return o.queryResult
}

func (o *LazyStartsWithOperation) setQueryResult(queryResult *QueryResult) {
	o.queryResult = queryResult
}

func (o *LazyStartsWithOperation) isRequiresRetry() bool {
	return o.requiresRetry
}

func (o *LazyStartsWithOperation) setRequiresRetry(requiresRetry bool) {
	o.requiresRetry = requiresRetry
}

func (o *LazyStartsWithOperation) handleResponse(response *GetResponse) error {
	var getDocumentResult *GetDocumentsResult
	err := jsonUnmarshal(response.result, &getDocumentResult)
	if err != nil {
		return err
	}

	finalResults := map[string]interface{}{}
	//TreeMap<string, Object> finalResults = new TreeMap<>(string::compareToIgnoreCase);

	for _, document := range getDocumentResult.Results {
		newDocumentInfo := getNewDocumentInfo(document)
		o._sessionOperations.documentsByID.add(newDocumentInfo)

		if newDocumentInfo.id == "" {
			continue // is this possible?
		}

		id := strings.ToLower(newDocumentInfo.id)
		if o._sessionOperations.IsDeleted(newDocumentInfo.id) {
			finalResults[id] = nil
			continue
		}
		doc := o._sessionOperations.documentsByID.getValue(newDocumentInfo.id)
		if doc != nil {
			finalResults[id], err = o._sessionOperations.TrackEntityInDocumentInfoOld(o._clazz, doc)
			if err != nil {
				return err
			}
			continue
		}
		finalResults[id] = nil
	}
	o.result = finalResults
	return nil
}
