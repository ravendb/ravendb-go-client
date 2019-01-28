package ravendb

import (
	"fmt"
	"reflect"
)

type LoadStartingWithOperation struct {
	_session *InMemoryDocumentSessionOperations

	_startWith  string
	_matches    string
	_start      int
	_pageSize   int
	_exclude    string
	_startAfter string

	_returnedIds []string

	Command *GetDocumentsCommand
}

func NewLoadStartingWithOperation(session *InMemoryDocumentSessionOperations) *LoadStartingWithOperation {
	return &LoadStartingWithOperation{
		_session: session,
	}
}

func (o *LoadStartingWithOperation) CreateRequest() (*GetDocumentsCommand, error) {
	if err := o._session.incrementRequestCount(); err != nil {
		return nil, err
	}

	var err error
	o.Command, err = NewGetDocumentsCommandFull(o._startWith, o._startAfter, o._matches, o._exclude, o._start, o._pageSize, false)
	return o.Command, err
}

func (o *LoadStartingWithOperation) withStartWith(idPrefix string) {
	o.withStartWithFull(idPrefix, "", 0, 0, "", "")
}

func (o *LoadStartingWithOperation) withStartWithAndMatches(idPrefix string, matches string) {
	o.withStartWithFull(idPrefix, matches, 0, 0, "", "")
}

func (o *LoadStartingWithOperation) withStartWithFull(idPrefix string, matches string, start int, pageSize int, exclude string, startAfter string) {
	o._startWith = idPrefix
	o._matches = matches
	o._start = start
	o._pageSize = pageSize
	o._exclude = exclude
	o._startAfter = startAfter
}

func (o *LoadStartingWithOperation) setResult(result *GetDocumentsResult) {
	documents := result.Results

	for _, document := range documents {
		newDocumentInfo := getNewDocumentInfo(document)
		o._session.documentsByID.add(newDocumentInfo)
		o._returnedIds = append(o._returnedIds, newDocumentInfo.id)
	}
}

// results must be *[]*struct. If *results is nil, we create it
func (o *LoadStartingWithOperation) getDocuments(results interface{}) error {
	//fmt.Printf("type of results: %T\n", results)
	rt := reflect.TypeOf(results)

	if rt.Kind() != reflect.Ptr || rt.Elem().Kind() != reflect.Slice {
		return fmt.Errorf("results should be a pointer to a slice of pointers to struct, is %T. rt: %s", results, rt)
	}
	rv := reflect.ValueOf(results)
	sliceV := rv.Elem()

	// slice element should be a pointer to a struct
	sliceElemPtrType := sliceV.Type().Elem()
	//fmt.Printf("type of sliceElemPtrType: %s\n", sliceElemPtrType.String())

	if sliceElemPtrType.Kind() != reflect.Ptr {
		return fmt.Errorf("results should be a pointer to a slice of pointers to struct, is %T. sliceElemPtrType: %s", results, sliceElemPtrType)
	}

	sliceElemType := sliceElemPtrType.Elem()
	if sliceElemType.Kind() != reflect.Struct {
		return fmt.Errorf("results should be a pointer to a slice of pointers to struct, is %T. sliceElemType: %s", results, sliceElemType)
	}
	// if this is a pointer to nil slice, create a new slice
	// otherwise we use the slice that was provided by the caller
	if sliceV.IsNil() {
		sliceV.Set(reflect.MakeSlice(sliceV.Type(), 0, 0))
	}

	sliceV2 := sliceV

	//resultType := reflect.PtrTo(sliceElemPtrType)
	//fmt.Printf("resultType: %s\n", resultType.String())
	for _, id := range o._returnedIds {
		rv := reflect.New(sliceElemPtrType)
		//fmt.Printf("type of rv: %T, %s\n", rv.Interface(), rv.Type().String())
		// rv is **struct and is set to value of *struct inside getDocument()
		err := o.getDocument(rv.Interface(), id)
		if err != nil {
			return err
		}
		// rv.Elem() is *struct
		sliceV2 = reflect.Append(sliceV2, rv.Elem())
	}

	if sliceV2 != sliceV {
		sliceV.Set(sliceV2)
	}
	return nil
}

func (o *LoadStartingWithOperation) getDocument(result interface{}, id string) error {
	// TODO: set to default value if not returning anything? Return ErrNotFound?
	if o._session.IsDeleted(id) {
		return nil
	}

	doc := o._session.documentsByID.getValue(id)
	if doc != nil {
		return o._session.TrackEntityInDocumentInfo(result, doc)
	}

	return nil
}
