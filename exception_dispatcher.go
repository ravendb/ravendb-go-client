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
	URL     string `json:"Url"`
	Type    string `json:"Type"`
	Message string `json:"Message"`
	Error   string `json:"Error"`
}

func NewExceptionSchema() *ExceptionSchema {
	return &ExceptionSchema{}
}

func (e *ExceptionSchema) getUrl() string {
	return e.URL
}

func (e *ExceptionSchema) setUrl(url string) {
	e.URL = url
}

func (e *ExceptionSchema) getType() string {
	return e.Type
}

func (e *ExceptionSchema) setType(typ string) {
	e.Type = typ
}

func (e *ExceptionSchema) getMessage() string {
	return e.Message
}

func (e *ExceptionSchema) setMessage(message string) {
	e.Message = message
}

func (e *ExceptionSchema) getError() string {
	return e.Error
}

func (e *ExceptionSchema) setError(err string) {
	e.Error = err
}
