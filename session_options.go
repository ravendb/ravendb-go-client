package ravendb

type SessionOptions struct {
	database        string
	requestExecutor *RequestExecutor
}

func NewSessionOptions() *SessionOptions {
	return &SessionOptions{}
}

func (o *SessionOptions) getDatabase() string {
	return o.database
}

func (o *SessionOptions) setDatabase(database string) {
	o.database = database
}

func (o *SessionOptions) getRequestExecutor() *RequestExecutor {
	return o.requestExecutor
}

func (o *SessionOptions) setRequestExecutor(requestExecutor *RequestExecutor) {
	o.requestExecutor = requestExecutor
}
