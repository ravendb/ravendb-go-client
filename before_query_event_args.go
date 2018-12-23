package ravendb

// BeforeQueryEventArgs describes arguments for "before query" event
type BeforeQueryEventArgs struct {
	Session            *InMemoryDocumentSessionOperations
	QueryCustomization *DocumentQueryCustomization
}
