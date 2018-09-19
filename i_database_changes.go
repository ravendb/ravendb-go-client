package ravendb

// IConnectableChanges is folded into IDatabaseChanges

type IDatabaseChanges interface {
	// those are IConnectableChanges
	isConnected() bool
	ensureConnectedNow() error
	addConnectionStatusChanged(handler func()) int
	removeConnectionStatusChanged(handlerIdx int)
	addOnError(handler func(error)) int
	removeOnError(handlerIdx int)

	forIndex(indexName string) (IChangesObservable, error)                      // *IndexChange
	forDocument(docID string) (IChangesObservable, error)                       // *DocumentChange>
	forAllDocuments() (IChangesObservable, error)                               // DocumentChange
	forOperationId(operationID int) (IChangesObservable, error)                 // OperationStatusChange
	forAllOperations() (IChangesObservable, error)                              // *OperationStatusChange
	forAllIndexes() (IChangesObservable, error)                                 // *IndexChange
	forDocumentsStartingWith(docIdPrefix string) (IChangesObservable, error)    // *DocumentChange>
	forDocumentsInCollection(collectionName string) (IChangesObservable, error) // *DocumentChange
	//IChangesObservable<DocumentChange> forDocumentsInCollection(Class<?> clazz);
	forDocumentsOfType(typeName string) (IChangesObservable, error) // *DocumentChange
	//IChangesObservable<DocumentChange> forDocumentsOfType(Class<?> clazz);
}
