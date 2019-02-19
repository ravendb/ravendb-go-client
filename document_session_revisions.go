package ravendb

// Note: Java's IRevisionsSessionOperations is DocumentSessionRevisions

// TODO: write a unique wrapper type
type RevisionsSessionOperations = DocumentSessionRevisions

// DocumentSessionRevisions represents revisions operations
type DocumentSessionRevisions struct {
	*AdvancedSessionExtensionBase
}

func newDocumentSessionRevisions(session *InMemoryDocumentSessionOperations) *DocumentSessionRevisions {
	return &DocumentSessionRevisions{
		AdvancedSessionExtensionBase: newAdvancedSessionExtensionBase(session),
	}
}

func (r *DocumentSessionRevisions) GetFor(results interface{}, id string) error {
	return r.GetForPaged(results, id, 0, 25)
}

func (r *DocumentSessionRevisions) GetForStartAt(results interface{}, id string, start int) error {
	return r.GetForPaged(results, id, start, 25)
}

func (r *DocumentSessionRevisions) GetForPaged(results interface{}, id string, start int, pageSize int) error {
	operation, err := NewGetRevisionOperationRange(r.session, id, start, pageSize, false)
	if err != nil {
		return err
	}

	command, err := operation.createRequest()
	if err != nil {
		return err
	}
	err = r.requestExecutor.ExecuteCommand(command, r.sessionInfo)
	if err != nil {
		return err
	}
	operation.setResult(command.Result)
	return operation.GetRevisionsFor(results)
}

func (r *DocumentSessionRevisions) GetMetadataFor(id string) ([]*MetadataAsDictionary, error) {
	return r.GetMetadataForPaged(id, 0, 25)
}

func (r *DocumentSessionRevisions) GetMetadataForStartAt(id string, start int) ([]*MetadataAsDictionary, error) {
	return r.GetMetadataForPaged(id, start, 25)
}

func (r *DocumentSessionRevisions) GetMetadataForPaged(id string, start int, pageSize int) ([]*MetadataAsDictionary, error) {
	operation, err := NewGetRevisionOperationRange(r.session, id, start, pageSize, true)
	if err != nil {
		return nil, err
	}
	command, err := operation.createRequest()
	if err != nil {
		return nil, err
	}
	err = r.requestExecutor.ExecuteCommand(command, r.sessionInfo)
	if err != nil {
		return nil, err
	}
	operation.setResult(command.Result)
	return operation.GetRevisionsMetadataFor(), nil
}

func (r *DocumentSessionRevisions) Get(result interface{}, changeVector string) error {
	operation := NewGetRevisionOperationWithChangeVectors(r.session, []string{changeVector})
	command, err := operation.createRequest()
	if err != nil {
		return err
	}
	err = r.requestExecutor.ExecuteCommand(command, r.sessionInfo)
	if err != nil {
		return err
	}
	operation.setResult(command.Result)
	return operation.GetRevision(result)
}

// TODO: needs tests
func (r *DocumentSessionRevisions) GetRevisions(results interface{}, changeVectors []string) error {
	operation := NewGetRevisionOperationWithChangeVectors(r.session, changeVectors);

	command, err := operation.createRequest();
	if err != nil {
		return err
	}
	err = r.requestExecutor.ExecuteCommand(command, r.sessionInfo);
	if err != nil {
		return err
	}
	operation.setResult(command.Result);
	return operation.GetRevisions(results);
}
