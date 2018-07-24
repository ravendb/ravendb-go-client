package ravendb

type MethodCall interface {
}

type MethodCallData struct {
	args       []Object
	accessPath string
}
