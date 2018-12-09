package ravendb

type BatchOperation struct {
	_session              *InMemoryDocumentSessionOperations
	_entities             []interface{}
	_sessionCommandsCount int
}

func NewBatchOperation(session *InMemoryDocumentSessionOperations) *BatchOperation {
	return &BatchOperation{
		_session: session,
	}
}

func (b *BatchOperation) CreateRequest() (*BatchCommand, error) {
	result, err := b._session.PrepareForSaveChanges()
	if err != nil {
		return nil, err
	}

	b._sessionCommandsCount = len(result.GetSessionCommands())
	result.sessionCommands = append(result.sessionCommands, result.GetDeferredCommands()...)
	if len(result.GetSessionCommands()) == 0 {
		return nil, nil
	}

	err = b._session.IncrementRequestCount()
	if err != nil {
		return nil, err
	}

	b._entities = result.GetEntities()

	return NewBatchCommandWithOptions(b._session.GetConventions(), result.GetSessionCommands(), result.GetOptions())
}

func (b *BatchOperation) setResult(result ArrayNode) {
	if len(result) == 0 {
		// TODO: throwOnNullResults()
		return
	}
	for i := 0; i < b._sessionCommandsCount; i++ {
		batchResult := result[i]
		if batchResult == nil {
			return
			//TODO: throw new IllegalArgumentException();
		}
		typ, _ := JsonGetAsText(batchResult, "Type")
		if typ != "PUT" {
			continue
		}
		entity := b._entities[i]
		documentInfo := b._session.documentsByEntity[entity]
		if documentInfo == nil {
			continue
		}
		changeVector := jsonGetAsTextPointer(batchResult, Constants_Documents_Metadata_CHANGE_VECTOR)
		if changeVector == nil {
			return
			//TODO: throw new IllegalStateException("PUT response is invalid. @change-vector is missing on " + documentInfo.GetId());
		}
		id, _ := JsonGetAsText(batchResult, Constants_Documents_Metadata_ID)
		if id == "" {
			return
			//TODO: throw new IllegalStateException("PUT response is invalid. @id is missing on " + documentInfo.GetId());
		}

		for propertyName, v := range batchResult {
			if propertyName == "Type" {
				continue
			}

			meta := documentInfo.metadata
			meta[propertyName] = v
		}

		documentInfo.id = id
		documentInfo.changeVector = changeVector
		doc := documentInfo.document
		doc[Constants_Documents_Metadata_KEY] = documentInfo.metadata
		documentInfo.metadataInstance = nil

		b._session.documentsById.add(documentInfo)
		b._session.GetGenerateEntityIdOnTheClient().trySetIdentity(entity, id)

		afterSaveChangesEventArgs := NewAfterSaveChangesEventArgs(b._session, documentInfo.id, documentInfo.entity)
		b._session.OnAfterSaveChangesInvoke(afterSaveChangesEventArgs)
	}
}
