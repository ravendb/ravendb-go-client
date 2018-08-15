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
func (o *GetRevisionOperation) GetRevisionWithDocument(clazz reflect.Type, document ObjectNode) interface{} {
	if document == nil {
		return Defaults_defaultValue(clazz)
	}

	var metadata ObjectNode
	id := ""
	if v, ok := document[Constants_Documents_Metadata_KEY]; ok {
		metadata = v.(ObjectNode)
		id, _ = jsonGetAsText(metadata, Constants_Documents_Metadata_ID)
	}
	var changeVector *string

	if metadata != nil {
		changeVector = jsonGetAsTextPointer(metadata, Constants_Documents_Metadata_CHANGE_VECTOR)
	}
	entity := o._session.getEntityToJson().convertToEntity(clazz, id, document)
	documentInfo := NewDocumentInfo()
	documentInfo.setId(id)
	documentInfo.setChangeVector(changeVector)
	documentInfo.setDocument(document)
	documentInfo.setMetadata(metadata)
	documentInfo.setEntity(entity)
	o._session.documentsByEntity[entity] = documentInfo
	return entity
}

func (o *GetRevisionOperation) GetRevisionsFor(clazz reflect.Type) []interface{} {
	resultsCount := len(o._result.getResults())
	results := make([]interface{}, resultsCount, resultsCount)
	for i := 0; i < resultsCount; i++ {
		document := o._result.getResults()[i]
		results[i] = o.GetRevisionWithDocument(clazz, document)
	}

	return results
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

func (o *GetRevisionOperation) GetRevision(clazz reflect.Type) interface{} {
	if o._result == nil {
		return Defaults_defaultValue(clazz)
	}

	document := o._result.getResults()[0]
	return o.GetRevisionWithDocument(clazz, document)
}

func (o *GetRevisionOperation) GetRevisions(clazz reflect.Type) map[string]interface{} {
	// Maybe: Java uses case-insensitive keys, but keys are change vectors
	// so that shouldn't matter
	results := map[string]interface{}{}

	for i := 0; i < len(o._command.GetChangeVectors()); i++ {
		changeVector := o._command.GetChangeVectors()[i]
		if changeVector == "" {
			continue
		}

		v := o._result.getResults()[i]
		rev := o.GetRevisionWithDocument(clazz, v)
		results[changeVector] = rev
	}

	return results
}
