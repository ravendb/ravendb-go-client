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
	results       interface{}
	queryResult   *QueryResult
	requiresRetry bool
}

// NewLazyStartsWithOperation returns new LazyStartsWithOperation
// TODO: convert to use StartsWithArgs
// TODO: validate that results is map[string]*Struct
// results is map[string]*Struct
func NewLazyStartsWithOperation(results interface{}, idPrefix string, matches string, exclude string, start int, pageSize int, sessionOperations *InMemoryDocumentSessionOperations, startAfter string) *LazyStartsWithOperation {
	return &LazyStartsWithOperation{
		idPrefix:          idPrefix,
		matches:           matches,
		exclude:           exclude,
		start:             start,
		pageSize:          pageSize,
		sessionOperations: sessionOperations,
		startAfter:        startAfter,
		results:           results,
	}
}

func (o *LazyStartsWithOperation) createRequest() *GetRequest {
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

	request := &GetRequest{
		url:   "/docs",
		query: q,
	}

	return request
}

// needed for ILazyOperation
func (o *LazyStartsWithOperation) getResult() interface{} {
	return o.results
}

// needed for ILazyOperation
func (o *LazyStartsWithOperation) getQueryResult() *QueryResult {
	return o.queryResult
}

// needed for ILazyOperation
func (o *LazyStartsWithOperation) isRequiresRetry() bool {
	return o.requiresRetry
}

func (o *LazyStartsWithOperation) handleResponse(response *GetResponse) error {
	var getDocumentResult *GetDocumentsResult
	err := jsonUnmarshal(response.result, &getDocumentResult)
	if err != nil {
		return err
	}

	for _, document := range getDocumentResult.Results {
		newDocumentInfo := getNewDocumentInfo(document)
		o.sessionOperations.documentsByID.add(newDocumentInfo)

		if newDocumentInfo.id == "" {
			continue // is this possible?
		}

		id := strings.ToLower(newDocumentInfo.id)

		var tp reflect.Type
		var ok bool
		if tp, ok = isMapStringToPtrStruct(reflect.TypeOf(o.results)); !ok {
			return fmt.Errorf("expected o.results to be of type map[string]*struct, got %T", o.results)
		}
		finalResult := reflect.ValueOf(o.results)
		key := reflect.ValueOf(id)
		if o.sessionOperations.IsDeleted(newDocumentInfo.id) {
			nilPtr := reflect.New(tp)
			finalResult.SetMapIndex(key, reflect.ValueOf(nilPtr))
			continue
		}
		doc := o.sessionOperations.documentsByID.getValue(newDocumentInfo.id)
		if doc != nil {
			v := reflect.New(tp).Interface()
			err = o.sessionOperations.TrackEntityInDocumentInfo(v, doc)
			if err != nil {
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
