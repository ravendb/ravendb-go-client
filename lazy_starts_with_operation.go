package ravendb

import (
	"fmt"
	"reflect"
	"strings"
)

var _ ILazyOperation = &LazyStartsWithOperation{}

// LazyStartsWithOperation represents lazy starts with operation
type LazyStartsWithOperation struct {
	idPrefix          string
	matches           string
	exclude           string
	start             int
	pageSize          int
	sessionOperations *InMemoryDocumentSessionOperations
	startAfter        string

	// results is map[string]*Struct
	queryResult   *QueryResult
	requiresRetry bool
	rawResult     *GetDocumentsResult
}

// NewLazyStartsWithOperation returns new LazyStartsWithOperation
// TODO: convert to use StartsWithArgs
func NewLazyStartsWithOperation(idPrefix string, matches string, exclude string, start int, pageSize int, sessionOperations *InMemoryDocumentSessionOperations, startAfter string) *LazyStartsWithOperation {
	return &LazyStartsWithOperation{
		idPrefix:          idPrefix,
		matches:           matches,
		exclude:           exclude,
		start:             start,
		pageSize:          pageSize,
		sessionOperations: sessionOperations,
		startAfter:        startAfter,
	}
}

// needed for ILazyOperation
func (o *LazyStartsWithOperation) createRequest() *getRequest {
	pageSize := o.pageSize
	if pageSize == 0 {
		pageSize = 25
	}
	q := fmt.Sprintf("?startsWith=%s&matches=%s&exclude=%s&start=%d&pageSize=%d&startAfter=%s",
		urlUtilsEscapeDataString(o.idPrefix),
		urlUtilsEscapeDataString(o.matches),
		urlUtilsEscapeDataString(o.exclude),
		o.start,
		pageSize,
		o.startAfter)

	request := &getRequest{
		url:   "/docs",
		query: q,
	}

	return request
}

// needed for ILazyOperation
// results should be map[string]*<type>
func (o *LazyStartsWithOperation) getResult(results interface{}) error {
	var tp reflect.Type
	var ok bool
	if tp, ok = isMapStringToPtrStruct(reflect.TypeOf(results)); !ok {
		return fmt.Errorf("expected o.results to be of type map[string]*struct, got %T", results)
	}

	finalResult := reflect.ValueOf(results)

	for _, document := range o.rawResult.Results {
		newDocumentInfo := getNewDocumentInfo(document)
		o.sessionOperations.documentsByID.add(newDocumentInfo)

		if newDocumentInfo.id == "" {
			continue // is this possible?
		}

		id := strings.ToLower(newDocumentInfo.id)

		key := reflect.ValueOf(id)
		if o.sessionOperations.IsDeleted(newDocumentInfo.id) {
			nilPtr := reflect.New(tp)
			finalResult.SetMapIndex(key, reflect.ValueOf(nilPtr))
			continue
		}
		doc := o.sessionOperations.documentsByID.getValue(newDocumentInfo.id)
		if doc != nil {
			v := reflect.New(tp).Interface()
			if err := o.sessionOperations.TrackEntityInDocumentInfo(v, doc); err != nil {
				return err
			}
			finalResult.SetMapIndex(key, reflect.ValueOf(v).Elem())
			continue
		}
		nilPtr := reflect.New(tp)
		finalResult.SetMapIndex(key, reflect.ValueOf(nilPtr))
	}
	return nil
}

// needed for ILazyOperation
func (o *LazyStartsWithOperation) getQueryResult() *QueryResult {
	return o.queryResult
}

// needed for ILazyOperation
func (o *LazyStartsWithOperation) isRequiresRetry() bool {
	return o.requiresRetry
}

// needed for ILazyOperation
func (o *LazyStartsWithOperation) handleResponse(response *GetResponse) error {
	var getDocumentResult *GetDocumentsResult
	if err := jsonUnmarshal(response.Result, &getDocumentResult); err != nil {
		return err
	}
	o.rawResult = getDocumentResult
	return nil
}
