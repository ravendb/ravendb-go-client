package ravendb

// BatchOperation represents a batch operation
type BatchOperation struct {
	session              *InMemoryDocumentSessionOperations
	entities             []interface{}
	sessionCommandsCount int
}

func newBatchOperation(session *InMemoryDocumentSessionOperations) *BatchOperation {
	return &BatchOperation{
		session: session,
	}
}

func (b *BatchOperation) createRequest() (*BatchCommand, error) {
	result, err := b.session.prepareForSaveChanges()
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

	return newBatchCommand(b.session.GetConventions(), result.sessionCommands, result.options, b.session.transactionMode, b.session.disableAtomicDocumentWritesInClusterWideTransaction)
}

func (b *BatchOperation) setResult(serverResult *JSONArrayResult) error {
	b.session.sessionInfo.lastClusterTransactionIndex = &serverResult.TransactionIndex
	if b.session.transactionMode == TransactionMode_ClusterWide && serverResult.TransactionIndex <= 0 {
		url, err := b.session.GetRequestExecutor().GetURL()
		if err != nil {
			url = "UNKOWN_URL"
		}
		return newIllegalStateError("Cluster transaction was send to a node that is not supporting it.  So it was executed ONLY on the requested node on " + url)
	}

	result := serverResult.Results
	if len(result) == 0 {
		return throwOnNullResult()
	}

	for commandIndex := 0; commandIndex < b.sessionCommandsCount; commandIndex++ {
		batchResult := result[commandIndex]
		if batchResult == nil {
			return newIllegalArgumentError("batchResult cannot be nil")
		}
		typ, _ := jsonGetAsText(batchResult, "Type")

		switch typ {
		case "PUT":
			entity := b.entities[commandIndex]
			documentInfo := getDocumentInfoByEntity(b.session.documentsByEntity, entity)
			if documentInfo == nil {
				continue
			}
			changeVector := jsonGetAsTextPointer(batchResult, MetadataChangeVector)
			if changeVector == nil {
				return newIllegalStateError("PUT response is invalid. @change-vector is missing on " + documentInfo.id)
			}
			id, _ := jsonGetAsText(batchResult, MetadataID)
			if id == "" {
				return newIllegalStateError("PUT response is invalid. @id is missing on " + documentInfo.id)
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
			b.session.generateEntityIDOnTheClient.trySetIdentity(entity, id)

			afterSaveChangesEventArgs := newAfterSaveChangesEventArgs(b.session, documentInfo.id, documentInfo.entity)
			b.session.onAfterSaveChangesInvoke(afterSaveChangesEventArgs)
			break
		case "CompareExchangePUT":
			index, exist := batchResult["Index"].(float64)
			if exist == false {
				return newIllegalStateError("CompareExchangePUT is missing index property.")
			}
			key, exist := batchResult["Key"].(string)
			if exist == false {
				return newIllegalStateError("CompareExchangePUT is missing key property.")
			}
			clusterSession, err := b.session.GetClusterSession()
			if err != nil {
				return err
			}

			clusterSession.updateState(key, int64(index))
			break
		default:
			break
		}

	}

	return nil
}

func (b *BatchOperation) setResultOld(result []map[string]interface{}) error {
	if len(result) == 0 {
		return throwOnNullResult()
	}
	for i := 0; i < b.sessionCommandsCount; i++ {
		batchResult := result[i]
		if batchResult == nil {
			return newIllegalArgumentError("batchResult cannot be nil")
		}
		typ, _ := jsonGetAsText(batchResult, "Type")
		if typ != "PUT" {
			continue
		}
		entity := b.entities[i]
		documentInfo := getDocumentInfoByEntity(b.session.documentsByEntity, entity)
		if documentInfo == nil {
			continue
		}
		changeVector := jsonGetAsTextPointer(batchResult, MetadataChangeVector)
		if changeVector == nil {
			return newIllegalStateError("PUT response is invalid. @change-vector is missing on " + documentInfo.id)
		}
		id, _ := jsonGetAsText(batchResult, MetadataID)
		if id == "" {
			return newIllegalStateError("PUT response is invalid. @id is missing on " + documentInfo.id)
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
		b.session.generateEntityIDOnTheClient.trySetIdentity(entity, id)

		afterSaveChangesEventArgs := newAfterSaveChangesEventArgs(b.session, documentInfo.id, documentInfo.entity)
		b.session.onAfterSaveChangesInvoke(afterSaveChangesEventArgs)
	}
	return nil
}

func throwOnNullResult() error {
	return newIllegalStateError("Received empty response from the server. This is not supposed to happen and is likely a bug.")
}
