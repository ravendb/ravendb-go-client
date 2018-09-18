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

	//IChangesObservable<IndexChange> forIndex(string indexName);
	//IChangesObservable<DocumentChange> forDocument(string docId);
	//IChangesObservable<DocumentChange> forAllDocuments();
	//IChangesObservable<OperationStatusChange> forOperationId(long operationId);
	forAllOperations() IChangesObservable // *OperationStatusChange
	//IChangesObservable<IndexChange> forAllIndexes();
	//IChangesObservable<DocumentChange> forDocumentsStartingWith(string docIdPrefix);
	//IChangesObservable<DocumentChange> forDocumentsInCollection(string collectionName);
	//IChangesObservable<DocumentChange> forDocumentsInCollection(Class<?> clazz);
	//IChangesObservable<DocumentChange> forDocumentsOfType(string typeName);
	//IChangesObservable<DocumentChange> forDocumentsOfType(Class<?> clazz);
}
