package ravendb

import (
	"fmt"
	"reflect"
	"strings"
)

// LoadOperation represents a load operation
type LoadOperation struct {
	session *InMemoryDocumentSessionOperations

	ids                []string
	includes           []string
	idsToCheckOnServer []string
}

func NewLoadOperation(session *InMemoryDocumentSessionOperations) *LoadOperation {
	return &LoadOperation{
		session: session,
	}
}

func (o *LoadOperation) createRequest() (*GetDocumentsCommand, error) {
	if len(o.idsToCheckOnServer) == 0 {
		return nil, nil
	}

	if o.session.checkIfIdAlreadyIncluded(o.ids, o.includes) {
		return nil, nil
	}

	if err := o.session.incrementRequestCount(); err != nil {
		return nil, err
	}

	return NewGetDocumentsCommand(o.idsToCheckOnServer, o.includes, false)
}

func (o *LoadOperation) byID(id string) *LoadOperation {
	if id == "" {
		return o
	}

	if o.ids == nil {
		o.ids = []string{id}
	}

	if o.session.IsLoadedOrDeleted(id) {
		return o
	}

	o.idsToCheckOnServer = append(o.idsToCheckOnServer, id)
	return o
}

func (o *LoadOperation) withIncludes(includes []string) *LoadOperation {
	o.includes = includes
	return o
}

func (o *LoadOperation) byIds(ids []string) *LoadOperation {
	o.ids = stringArrayCopy(ids)

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
		o.byID(id)
	}
	return o
}

func (o *LoadOperation) getDocument(result interface{}) error {
	return o.getDocumentWithID(result, o.ids[0])
}

func (o *LoadOperation) getDocumentWithID(result interface{}, id string) error {
	if id == "" {
		// TODO: should return default value?
		//return ErrNotFound
		return nil
	}

	if o.session.IsDeleted(id) {
		// TODO: return ErrDeleted?
		//return ErrNotFound
		return nil
	}

	doc := o.session.documentsByID.getValue(id)
	if doc == nil {
		doc = o.session.includedDocumentsByID[id]
	}
	if doc == nil {
		//return ErrNotFound
		return nil
	}

	return o.session.TrackEntityInDocumentInfo(result, doc)
}

var stringType = reflect.TypeOf("")

// TODO: also handle a pointer to a map?
func (o *LoadOperation) getDocuments(results interface{}) error {
	// results must be map[string]*struct
	//fmt.Printf("LoadOperation.getDocuments: results type: %T\n", results)
	m := reflect.ValueOf(results)
	if m.Type().Kind() != reflect.Map {
		return fmt.Errorf("results should be a map[string]*struct, is %s. tp: %s", m.Type().String(), m.Type().String())
	}
	mapKeyType := m.Type().Key()
	if mapKeyType != stringType {
		return fmt.Errorf("results should be a map[string]*struct, is %s. tp: %s", m.Type().String(), m.Type().String())
	}
	mapElemPtrType := m.Type().Elem()
	if mapElemPtrType.Kind() != reflect.Ptr {
		return fmt.Errorf("results should be a map[string]*struct, is %s. tp: %s", m.Type().String(), m.Type().String())
	}
	mapElemType := mapElemPtrType.Elem()
	if mapElemType.Kind() != reflect.Struct {
		return fmt.Errorf("results should be a map[string]*struct, is %s. tp: %s", m.Type().String(), m.Type().String())
	}

	uniqueIds := stringArrayCopy(o.ids)
	stringArrayRemove(&uniqueIds, "")
	uniqueIds = stringArrayRemoveDuplicatesNoCase(uniqueIds)
	for _, id := range uniqueIds {
		v := reflect.New(mapElemPtrType).Interface()
		err := o.getDocumentWithID(v, id)
		if err != nil {
			return err
		}
		key := reflect.ValueOf(id)
		v2 := reflect.ValueOf(v).Elem() // convert *<type> to <type>
		m.SetMapIndex(key, v2)
	}

	return nil
}

func (o *LoadOperation) setResult(result *GetDocumentsResult) {
	if result == nil {
		return
	}

	o.session.registerIncludes(result.Includes)

	results := result.Results
	for _, document := range results {
		// TODO: Java also does document.isNull()
		if document == nil {
			continue
		}
		newDocumentInfo := getNewDocumentInfo(document)
		o.session.documentsByID.add(newDocumentInfo)
	}

	o.session.registerMissingIncludes(result.Results, result.Includes, o.includes)
}
