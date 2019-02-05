package ravendb

// BatchOperation represents a batch operation
type BatchOperation struct {
	session              *InMemoryDocumentSessionOperations
	entities             []interface{}
	sessionCommandsCount int
}

// NewBatchOperation
func NewBatchOperation(session *InMemoryDocumentSessionOperations) *BatchOperation {
	return &BatchOperation{
		session: session,
	}
}

func (b *BatchOperation) createRequest() (*BatchCommand, error) {
	result, err := b.session.PrepareForSaveChanges()
	if err != nil {
		return nil, err
	}

	b.sessionCommandsCount = len(result.sessionCommands)
	result.sessionCommands = append(result.sessionCommands, result.deferredCommands...)
	if len(result.sessionCommands) == 0 {
		return nil, nil
	}

	if err = b.session.incrementRequestCount(); err != nil {
		return nil, err
	}

	b.entities = result.entities

	return NewBatchCommand(b.session.GetConventions(), result.sessionCommands, result.options)
}

func (b *BatchOperation) setResult(result []map[string]interface{}) {
	if len(result) == 0 {
		// TODO: throwOnNullResults()
		return
	}
	for i := 0; i < b.sessionCommandsCount; i++ {
		batchResult := result[i]
		if batchResult == nil {
			return
			//TODO: throw new IllegalArgumentError();
		}
		typ, _ := jsonGetAsText(batchResult, "Type")
		if typ != "PUT" {
			continue
		}
		entity := b.entities[i]
		documentInfo := getDocumentInfoByEntity(b.session.documents, entity)
		if documentInfo == nil {
			continue
		}
		changeVector := jsonGetAsTextPointer(batchResult, MetadataChangeVector)
		if changeVector == nil {
			return
			//TODO: throw new IllegalStateError("PUT response is invalid. @change-vector is missing on " + documentInfo.GetID());
		}
		id, _ := jsonGetAsText(batchResult, MetadataID)
		if id == "" {
			return
			//TODO: throw new IllegalStateError("PUT response is invalid. @id is missing on " + documentInfo.GetID());
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
		doc[MetadataKey] = documentInfo.metadata
		documentInfo.metadataInstance = nil

		b.session.documentsByID.add(documentInfo)
		b.session.GetGenerateEntityIDOnTheClient().trySetIdentity(entity, id)

		afterSaveChangesEventArgs := newAfterSaveChangesEventArgs(b.session, documentInfo.id, documentInfo.entity)
		b.session.OnAfterSaveChangesInvoke(afterSaveChangesEventArgs)
	}
}
