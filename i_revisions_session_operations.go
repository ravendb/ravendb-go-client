package ravendb

// Note: DocumentSessionRevisions is the only implementation of IRevisionsSessionOperations
// so for simplicty we fuse them
type IRevisionsSessionOperations = DocumentSessionRevisions
