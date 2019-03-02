package ravendb

import (
	"reflect"
)

// GetRevisionOperation represents "get revisions" operation
type GetRevisionOperation struct {
	session *InMemoryDocumentSessionOperations

	result  *JSONArrayResult
	command *GetRevisionsCommand
}

func NewGetRevisionOperationWithChangeVectors(session *InMemoryDocumentSessionOperations, changeVectors []string) *GetRevisionOperation {
	return &GetRevisionOperation{
		session: session,
		command: NewGetRevisionsCommand(changeVectors, false),
	}
}

func NewGetRevisionOperationRange(session *InMemoryDocumentSessionOperations, id string, start int, pageSize int, metadataOnly bool) (*GetRevisionOperation, error) {
	if session == nil {
		return nil, newIllegalArgumentError("session cannot be null")
	}
	if id == "" {
		return nil, newIllegalArgumentError("Id cannot be null")
	}
	return &GetRevisionOperation{
		session: session,
		command: NewGetRevisionsCommandRange(id, start, pageSize, metadataOnly),
	}, nil
}

func (o *GetRevisionOperation) createRequest() (*GetRevisionsCommand, error) {
	return o.command, nil
}

func (o *GetRevisionOperation) setResult(result *JSONArrayResult) {
	o.result = result
}

// Note: in Java it's getRevision
func (o *GetRevisionOperation) GetRevisionWithDocument(result interface{}, document map[string]interface{}) error {
	if document == nil {
		return nil
	}

	var metadata map[string]interface{}
	id := ""
	if v, ok := document[MetadataKey]; ok {
		metadata = v.(map[string]interface{})
		id, _ = jsonGetAsText(metadata, MetadataID)
	}
	var changeVector *string

	if metadata != nil {
		changeVector = jsonGetAsTextPointer(metadata, MetadataChangeVector)
	}
	err := o.session.getEntityToJSON().convertToEntity2(result, id, document)
	if err != nil {
		return err
	}
	documentInfo := &documentInfo{}
	documentInfo.id = id
	documentInfo.changeVector = changeVector
	documentInfo.document = document
	documentInfo.metadata = metadata
	documentInfo.setEntity(result)
	setDocumentInfo(&o.session.documentsByEntity, documentInfo)
	return nil
}

// results should be *[]*<type>
func (o *GetRevisionOperation) GetRevisionsFor(results interface{}) error {

	a := o.result.getResults()
	if len(a) == 0 {
		return nil
	}
	// TODO: optimize by creating pre-allocated slice of size len(a)
	slice, err := makeSliceForResults(results)
	if err != nil {
		return err
	}
	sliceElemType := reflect.TypeOf(results).Elem().Elem()

	tmpSlice := slice
	for _, document := range a {
		// creates a pointer to value e.g. **Foo
		resultV := reflect.New(sliceElemType)
		err = o.GetRevisionWithDocument(resultV.Interface(), document)
		if err != nil {
			return err
		}
		// append *Foo to the slice
		tmpSlice = reflect.Append(tmpSlice, resultV.Elem())
	}

	if slice != tmpSlice {
		slice.Set(tmpSlice)
	}
	return nil
}

func (o *GetRevisionOperation) GetRevisionsMetadataFor() []*MetadataAsDictionary {
	resultsCount := len(o.result.getResults())
	results := make([]*MetadataAsDictionary, resultsCount)
	for i := 0; i < resultsCount; i++ {
		document := o.result.getResults()[i]

		var metadata map[string]interface{}
		if v, ok := document[MetadataKey]; ok {
			metadata = v.(map[string]interface{})
		}
		results[i] = NewMetadataAsDictionaryWithSource(metadata)
	}
	return results
}

func (o *GetRevisionOperation) GetRevision(result interface{}) error {
	if o.result == nil {
		return nil
	}

	document := o.result.getResults()[0]
	return o.GetRevisionWithDocument(result, document)
}

// result should be map[string]<type>
func (o *GetRevisionOperation) GetRevisions(results interface{}) error {
	// Maybe: Java uses case-insensitive keys, but keys are change vectors
	// so that shouldn't matter

	rv := reflect.ValueOf(results)
	elemType, ok := isMapStringToPtrStruct(rv.Type())
	if !ok {
		return newIllegalArgumentError("results should be of type map[string]*<struct>, is %T", results)
	}

	for i := 0; i < len(o.command.GetChangeVectors()); i++ {
		changeVector := o.command.GetChangeVectors()[i]
		if changeVector == "" {
			continue
		}

		v := o.result.getResults()[i]
		// resultV is **Foo
		resultV := reflect.New(elemType)
		err := o.GetRevisionWithDocument(resultV.Interface(), v)
		if err != nil {
			return err
		}
		key := reflect.ValueOf(changeVector)
		rv.SetMapIndex(key, resultV.Elem())
	}

	return nil
}
