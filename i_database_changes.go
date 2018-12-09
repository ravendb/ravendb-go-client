package ravendb

// IConnectableChanges is folded into IDatabaseChanges

// TODO: remove IDatabaseChanges and make it just DatabaseChanges
// TODO: reduce surface of IConnectableChanges or remove completely
type IDatabaseChanges interface {
	// those are IConnectableChanges
	IsConnected() bool
	EnsureConnectedNow() error
	AddConnectionStatusChanged(handler func()) int
	RemoveConnectionStatusChanged(handlerIdx int)
	AddOnError(handler func(error)) int
	RemoveOnError(handlerIdx int)

	ForIndex(indexName string) (IChangesObservable, error)                      // *IndexChange
	ForDocument(docID string) (IChangesObservable, error)                       // *DocumentChange>
	ForAllDocuments() (IChangesObservable, error)                               // DocumentChange
	ForOperationId(operationID int) (IChangesObservable, error)                 // OperationStatusChange
	ForAllOperations() (IChangesObservable, error)                              // *OperationStatusChange
	ForAllIndexes() (IChangesObservable, error)                                 // *IndexChange
	ForDocumentsStartingWith(docIdPrefix string) (IChangesObservable, error)    // *DocumentChange>
	ForDocumentsInCollection(collectionName string) (IChangesObservable, error) // *DocumentChange
	//IChangesObservable<DocumentChange> forDocumentsInCollection(Class<?> clazz);
	ForDocumentsOfType(typeName string) (IChangesObservable, error) // *DocumentChange
	//IChangesObservable<DocumentChange> forDocumentsOfType(Class<?> clazz);

	Close()
}
