package ravendb

import "reflect"

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

func (o *LoadStartingWithOperation) CreateRequest() *GetDocumentsCommand {
	// TODO: should propagate error
	o._session.IncrementRequestCount()

	o.Command = NewGetDocumentsCommandFull(o._startWith, o._startAfter, o._matches, o._exclude, o._start, o._pageSize, false)
	return o.Command
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
	documents := result.GetResults()

	for _, document := range documents {
		newDocumentInfo := DocumentInfo_getNewDocumentInfo(document)
		o._session.documentsById.add(newDocumentInfo)
		o._returnedIds = append(o._returnedIds, newDocumentInfo.getId())
	}
}

func (o *LoadStartingWithOperation) getDocumentsOld(clazz reflect.Type) ([]interface{}, error) {
	var finalResults []interface{}
	for _, id := range o._returnedIds {
		v, err := o.getDocumentOld(clazz, id)
		if err != nil {
			return nil, err
		}
		finalResults = append(finalResults, v)
	}

	return finalResults, nil
}

func (o *LoadStartingWithOperation) getDocumentOld(clazz reflect.Type, id string) (interface{}, error) {
	if id == "" {
		return Defaults_defaultValue(clazz), nil
	}

	if o._session.IsDeleted(id) {
		return Defaults_defaultValue(clazz), nil
	}

	doc := o._session.documentsById.getValue(id)
	if doc != nil {
		return o._session.TrackEntityInDocumentInfoOld(clazz, doc)
	}

	return Defaults_defaultValue(clazz), nil
}
