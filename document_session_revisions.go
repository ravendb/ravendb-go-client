package ravendb

import "reflect"

type DocumentSessionRevisions struct {
	*AdvancedSessionExtentionBase
}

func NewDocumentSessionRevisions(session *InMemoryDocumentSessionOperations) *DocumentSessionRevisions {
	return &DocumentSessionRevisions{
		AdvancedSessionExtentionBase: NewAdvancedSessionExtentionBase(session),
	}
}

func (r *DocumentSessionRevisions) getFor(clazz reflect.Type, id string) ([]interface{}, error) {
	return r.getForPaged(clazz, id, 0, 25)
}

func (r *DocumentSessionRevisions) getForStartAt(clazz reflect.Type, id string, start int) ([]interface{}, error) {
	return r.getForPaged(clazz, id, start, 25)
}

// use -1 for start and pageSize to mean: "not given"
func (r *DocumentSessionRevisions) getForPaged(clazz reflect.Type, id string, start int, pageSize int) ([]interface{}, error) {
	operation := NewGetRevisionOperationRange(r.session, id, start, pageSize, false)

	command := operation.createRequest()
	err := r.requestExecutor.executeCommandWithSessionInfo(command, r.sessionInfo)
	if err != nil {
		return nil, err
	}
	operation.setResult(command.Result)
	return operation.getRevisionsFor(clazz), nil
}

/*
   public List<MetadataAsDictionary> getMetadataFor(String id) {
       return getMetadataFor(id, 0, 25);
   }

   public List<MetadataAsDictionary> getMetadataFor(String id, int start) {
       return getMetadataFor(id, start, 25);
   }

   public List<MetadataAsDictionary> getMetadataFor(String id, int start, int pageSize) {
       GetRevisionOperation operation = new GetRevisionOperation(session, id, start, pageSize, true);
       GetRevisionsCommand command = operation.createRequest();
       requestExecutor.execute(command, sessionInfo);
       operation.setResult(command.getResult());
       return operation.getRevisionsMetadataFor();
   }

   public <T> T get(Class<T> clazz, String changeVector) {
       GetRevisionOperation operation = new GetRevisionOperation(session, changeVector);

       GetRevisionsCommand command = operation.createRequest();
       requestExecutor.execute(command, sessionInfo);
       operation.setResult(command.getResult());
       return operation.getRevision(clazz);
   }

   public <T> Map<String, T> get(Class<T> clazz, String[] changeVectors) {
       GetRevisionOperation operation = new GetRevisionOperation(session, changeVectors);

       GetRevisionsCommand command = operation.createRequest();
       requestExecutor.execute(command, sessionInfo);
       operation.setResult(command.getResult());
       return operation.getRevisions(clazz);
   }
*/
