package ravendb

type SessionCreatedEventArgs struct {
	session *InMemoryDocumentSessionOperations
}

func NewSessionCreatedEventArgs(session *InMemoryDocumentSessionOperations) *SessionCreatedEventArgs {
	return &SessionCreatedEventArgs{
		session: session,
	}
}

func (a *SessionCreatedEventArgs) getSession() *InMemoryDocumentSessionOperations {
	return a.session
}
