package ravendb

import "reflect"

type IRevisionsSessionOperations = DocumentSessionRevisions

type DocumentSessionRevisions struct {
	*AdvancedSessionExtentionBase
}

func NewDocumentSessionRevisions(session *InMemoryDocumentSessionOperations) *DocumentSessionRevisions {
	return &DocumentSessionRevisions{
		AdvancedSessionExtentionBase: NewAdvancedSessionExtentionBase(session),
	}
}

func (r *DocumentSessionRevisions) GetFor(clazz reflect.Type, id string) ([]interface{}, error) {
	return r.GetForPaged(clazz, id, 0, 25)
}

func (r *DocumentSessionRevisions) GetForStartAt(clazz reflect.Type, id string, start int) ([]interface{}, error) {
	return r.GetForPaged(clazz, id, start, 25)
}

// use -1 for start and pageSize to mean: "not given"
func (r *DocumentSessionRevisions) GetForPaged(clazz reflect.Type, id string, start int, pageSize int) ([]interface{}, error) {
	operation := NewGetRevisionOperationRange(r.session, id, start, pageSize, false)

	command := operation.CreateRequest()
	err := r.requestExecutor.ExecuteCommandWithSessionInfo(command, r.sessionInfo)
	if err != nil {
		return nil, err
	}
	operation.setResult(command.Result)
	return operation.GetRevisionsFor(clazz)
}

func (r *DocumentSessionRevisions) GetMetadataFor(id string) ([]*MetadataAsDictionary, error) {
	return r.GetMetadataForPaged(id, 0, 25)
}

func (r *DocumentSessionRevisions) GetMetadataForStartAt(id string, start int) ([]*MetadataAsDictionary, error) {
	return r.GetMetadataForPaged(id, start, 25)
}

func (r *DocumentSessionRevisions) GetMetadataForPaged(id string, start int, pageSize int) ([]*MetadataAsDictionary, error) {
	operation := NewGetRevisionOperationRange(r.session, id, start, pageSize, true)
	command := operation.CreateRequest()
	err := r.requestExecutor.ExecuteCommandWithSessionInfo(command, r.sessionInfo)
	if err != nil {
		return nil, err
	}
	operation.setResult(command.Result)
	return operation.GetRevisionsMetadataFor(), nil
}

// TODO: change to take interface{} to return as an argument?
// TODO: change changeVector to *string?
func (r *DocumentSessionRevisions) Get(clazz reflect.Type, changeVector string) (interface{}, error) {
	operation := NewGetRevisionOperationWithChangeVector(r.session, changeVector)
	command := operation.CreateRequest()
	err := r.requestExecutor.ExecuteCommandWithSessionInfo(command, r.sessionInfo)
	if err != nil {
		return nil, err
	}
	operation.setResult(command.Result)
	return operation.GetRevision(clazz)
}

/*

   public <T> Map<String, T> get(Class<T> clazz, String[] changeVectors) {
       GetRevisionOperation operation = new GetRevisionOperation(Session, changeVectors);

       GetRevisionsCommand command = operation.CreateRequest();
       requestExecutor.execute(command, sessionInfo);
       operation.setResult(command.getResult());
       return operation.GetRevisions(clazz);
   }
*/
