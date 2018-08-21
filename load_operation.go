package ravendb

import (
	"reflect"
	"strings"
)

type LoadOperation struct {
	_session *InMemoryDocumentSessionOperations

	_ids                []string
	_includes           []string
	_idsToCheckOnServer []string
}

func NewLoadOperation(_session *InMemoryDocumentSessionOperations) *LoadOperation {
	return &LoadOperation{
		_session: _session,
	}
}

func (o *LoadOperation) CreateRequest() *GetDocumentsCommand {
	if len(o._idsToCheckOnServer) == 0 {
		return nil
	}

	if o._session.checkIfIdAlreadyIncluded(o._ids, o._includes) {
		return nil
	}

	// TODO: should propagate error
	o._session.IncrementRequestCount()

	return NewGetDocumentsCommand(o._idsToCheckOnServer, o._includes, false)
}

func (o *LoadOperation) byId(id string) *LoadOperation {
	if id == "" {
		return o
	}

	if o._ids == nil {
		o._ids = []string{id}
	}

	if o._session.IsLoadedOrDeleted(id) {
		return o
	}

	o._idsToCheckOnServer = append(o._idsToCheckOnServer, id)
	return o
}

func (o *LoadOperation) withIncludes(includes []string) *LoadOperation {
	o._includes = includes
	return o
}

func (o *LoadOperation) byIds(ids []string) *LoadOperation {
	o._ids = StringArrayCopy(ids)

	seen := map[string]struct{}{}
	for _, id := range ids {
		if id == "" {
			continue
		}
		idl := strings.ToLower(id)
		if _, ok := seen[idl]; ok {
			continue
		}
		seen[idl] = struct{}{}
		o.byId(id)
	}
	return o
}

func (o *LoadOperation) getDocument(clazz reflect.Type) (interface{}, error) {
	return o.getDocumentWithID(clazz, o._ids[0])
}

func (o *LoadOperation) getDocumentWithID(clazz reflect.Type, id string) (interface{}, error) {
	if id == "" {
		return Defaults_defaultValue(clazz), nil
	}

	if o._session.IsDeleted(id) {
		return Defaults_defaultValue(clazz), nil
	}

	doc := o._session.documentsById.getValue(id)
	if doc != nil {
		res, err := o._session.TrackEntityInDocumentInfo(clazz, doc)
		return res, err
	}

	doc, _ = o._session.includedDocumentsById[id]
	if doc != nil {
		res, err := o._session.TrackEntityInDocumentInfo(clazz, doc)
		return res, err
	}

	return Defaults_defaultValue(clazz), nil
}

func (o *LoadOperation) getDocuments(clazz reflect.Type) (map[string]interface{}, error) {
	uniqueIds := StringArrayCopy(o._ids)
	StringArrayRemove(&uniqueIds, "")
	uniqueIds = StringArrayRemoveDuplicatesNoCase(uniqueIds)
	res := make(map[string]interface{})
	for _, id := range uniqueIds {
		v, err := o.getDocumentWithID(clazz, id)
		if err != nil {
			return res, err
		}
		res[id] = v
	}
	return res, nil
}

func (o *LoadOperation) setResult(result *GetDocumentsResult) {
	if result == nil {
		return
	}

	o._session.RegisterIncludes(result.GetIncludes())

	results := result.GetResults()
	for _, document := range results {
		// TODO: Java also does document.isNull()
		if document == nil {
			continue
		}
		newDocumentInfo := DocumentInfo_getNewDocumentInfo(document)
		o._session.documentsById.add(newDocumentInfo)
	}

	o._session.RegisterMissingIncludes(result.GetResults(), result.GetIncludes(), o._includes)
}
