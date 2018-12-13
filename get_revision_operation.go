package ravendb

import "reflect"

type GetRevisionOperation struct {
	_session *InMemoryDocumentSessionOperations

	_result  *JSONArrayResult
	_command *GetRevisionsCommand
}

func NewGetRevisionOperationWithChangeVector(session *InMemoryDocumentSessionOperations, changeVector string) *GetRevisionOperation {
	return &GetRevisionOperation{
		_session: session,
		_command: NewGetRevisionsCommand([]string{changeVector}, false),
	}
}

func NewGetRevisionOperationRange(session *InMemoryDocumentSessionOperations, id string, start int, pageSize int, metadataOnly bool) *GetRevisionOperation {
	panicIf(session == nil, "Session cannot be null")
	panicIf(id == "", "Id cannot be null")
	return &GetRevisionOperation{
		_session: session,
		_command: NewGetRevisionsCommandRange(id, start, pageSize, metadataOnly),
	}
}

func (o *GetRevisionOperation) CreateRequest() *GetRevisionsCommand {
	return o._command
}

func (o *GetRevisionOperation) setResult(result *JSONArrayResult) {
	o._result = result
}

// Note: in Java it's getRevision
func (o *GetRevisionOperation) GetRevisionWithDocument(clazz reflect.Type, document ObjectNode) (interface{}, error) {
	if document == nil {
		return Defaults_defaultValue(clazz), nil
	}

	var metadata ObjectNode
	id := ""
	if v, ok := document[Constants_Documents_Metadata_KEY]; ok {
		metadata = v.(ObjectNode)
		id, _ = JsonGetAsText(metadata, Constants_Documents_Metadata_ID)
	}
	var changeVector *string

	if metadata != nil {
		changeVector = jsonGetAsTextPointer(metadata, Constants_Documents_Metadata_CHANGE_VECTOR)
	}
	entity, err := o._session.GetEntityToJson().ConvertToEntity(clazz, id, document)
	if err != nil {
		return nil, err
	}
	documentInfo := NewDocumentInfo()
	documentInfo.id = id
	documentInfo.changeVector = changeVector
	documentInfo.document = document
	documentInfo.metadata = metadata
	documentInfo.entity = entity
	setDocumentInfo(&o._session.documents, documentInfo)
	return entity, nil
}

func (o *GetRevisionOperation) GetRevisionsFor(clazz reflect.Type) ([]interface{}, error) {
	resultsCount := len(o._result.getResults())
	results := make([]interface{}, resultsCount, resultsCount)
	var err error
	for i := 0; i < resultsCount; i++ {
		document := o._result.getResults()[i]
		results[i], err = o.GetRevisionWithDocument(clazz, document)
		if err != nil {
			return nil, err
		}
	}

	return results, nil
}

func (o *GetRevisionOperation) GetRevisionsMetadataFor() []*MetadataAsDictionary {
	resultsCount := len(o._result.getResults())
	results := make([]*MetadataAsDictionary, resultsCount, resultsCount)
	for i := 0; i < resultsCount; i++ {
		document := o._result.getResults()[i]

		var metadata ObjectNode
		if v, ok := document[Constants_Documents_Metadata_KEY]; ok {
			metadata = v.(ObjectNode)
		}
		results[i] = NewMetadataAsDictionaryWithSource(metadata)
	}
	return results
}

func (o *GetRevisionOperation) GetRevision(clazz reflect.Type) (interface{}, error) {
	if o._result == nil {
		return Defaults_defaultValue(clazz), nil
	}

	document := o._result.getResults()[0]
	return o.GetRevisionWithDocument(clazz, document)
}

func (o *GetRevisionOperation) GetRevisions(clazz reflect.Type) (map[string]interface{}, error) {
	// Maybe: Java uses case-insensitive keys, but keys are change vectors
	// so that shouldn't matter
	results := map[string]interface{}{}

	for i := 0; i < len(o._command.GetChangeVectors()); i++ {
		changeVector := o._command.GetChangeVectors()[i]
		if changeVector == "" {
			continue
		}

		v := o._result.getResults()[i]
		rev, err := o.GetRevisionWithDocument(clazz, v)
		if err != nil {
			return nil, err
		}
		results[changeVector] = rev
	}

	return results, nil
}
