package ravendb

type MethodCall interface {
}

type MethodCallData struct {
	args       []interface{}
	accessPath string
}
