package ravendb

type BeforeQueryEventArgs struct {
	session            *InMemoryDocumentSessionOperations
	queryCustomization *DocumentQueryCustomization
}

func NewBeforeQueryEventArgs(session *InMemoryDocumentSessionOperations, queryCustomization *DocumentQueryCustomization) *BeforeQueryEventArgs {
	return &BeforeQueryEventArgs{
		session:            session,
		queryCustomization: queryCustomization,
	}
}

func (a *BeforeQueryEventArgs) getSession() *InMemoryDocumentSessionOperations {
	return a.session
}

func (a *BeforeQueryEventArgs) getQueryCustomization() *DocumentQueryCustomization {
	return a.queryCustomization
}
