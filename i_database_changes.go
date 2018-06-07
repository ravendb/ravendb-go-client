package ravendb

type IDatabaseChanges interface {
	// those are IConnectableChanges
	isConnected() bool
	ensureConnectedNow()
	addConnectionStatusChanged(handler EventHandler)
	removeConnectionStatusChanged(handler EventHandler)
	addOnError(handler Consumer)
	removeOnError(handler Consumer)

	//IChangesObservable<IndexChange> forIndex(String indexName);
	//IChangesObservable<DocumentChange> forDocument(String docId);
	//IChangesObservable<DocumentChange> forAllDocuments();
	//IChangesObservable<OperationStatusChange> forOperationId(long operationId);
	forAllOperations() IChangesObservable_OperationStatusChange
	//IChangesObservable<IndexChange> forAllIndexes();
	//IChangesObservable<DocumentChange> forDocumentsStartingWith(String docIdPrefix);
	//IChangesObservable<DocumentChange> forDocumentsInCollection(String collectionName);
	//IChangesObservable<DocumentChange> forDocumentsInCollection(Class<?> clazz);
	//IChangesObservable<DocumentChange> forDocumentsOfType(String typeName);
	//IChangesObservable<DocumentChange> forDocumentsOfType(Class<?> clazz);
}
