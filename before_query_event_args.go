package ravendb

type BeforeQueryEventArgs struct {
	session            *InMemoryDocumentSessionOperations
	queryCustomization *IDocumentQueryCustomization
}

func NewBeforeQueryEventArgs(session *InMemoryDocumentSessionOperations, queryCustomization *IDocumentQueryCustomization) *BeforeQueryEventArgs {
	return &BeforeQueryEventArgs{
		session:            session,
		queryCustomization: queryCustomization,
	}
}

func (a *BeforeQueryEventArgs) getSession() *InMemoryDocumentSessionOperations {
	return a.session
}

func (a *BeforeQueryEventArgs) getQueryCustomization() *IDocumentQueryCustomization {
	return a.queryCustomization
}
