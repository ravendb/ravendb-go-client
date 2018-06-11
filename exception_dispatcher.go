package ravendb

import "net/http"

func ExceptionDispatcher_get(schema *ExceptionSchema, code int) error {
	// TODO: Java is more complicated
	return ExceptionDispatcher_get2(schema.getMessage(), schema.getError(), schema.getType(), code)
}

func ExceptionDispatcher_get2(message string, err string, typeAsString string, code int) error {
	// TODO: Java is more complicated
	return NewRavenException("ExceptionDispatcher_get: http response exception")
}

func ExceptionDispatcher_throwException(response *http.Response) error {
	// TODO: Java is more complicated
	if response.Body != nil {
		response.Body.Close()
	}
	return NewRavenException("ExceptionDispatcher_get: http response exception")
}

type ExceptionSchema struct {
	Url     string `json:"Url"`
	Type    string `json:"Type"`
	Message string `json"Message"`
	Error   string `json:"Error"`
}

func NewExceptionSchema() *ExceptionSchema {
	return &ExceptionSchema{}
}

func (e *ExceptionSchema) getUrl() String {
	return e.Url
}

func (e *ExceptionSchema) setUrl(url String) {
	e.Url = url
}

func (e *ExceptionSchema) getType() String {
	return e.Type
}

func (e *ExceptionSchema) setType(typ String) {
	e.Type = typ
}

func (e *ExceptionSchema) getMessage() String {
	return e.Message
}

func (e *ExceptionSchema) setMessage(message String) {
	e.Message = message
}

func (e *ExceptionSchema) getError() String {
	return e.Error
}

func (e *ExceptionSchema) setError(err String) {
	e.Error = err
}
