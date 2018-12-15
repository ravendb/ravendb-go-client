package ravendb

import (
	"io/ioutil"
	"net/http"
	"strings"
)

func ExceptionDispatcher_get(schema *ExceptionSchema, code int) error {
	return ExceptionDispatcher_get2(schema.getMessage(), schema.getError(), schema.getType(), code)
}

func ExceptionDispatcher_get2(message string, err string, typeAsString string, code int) error {
	if code == http.StatusConflict {
		if strings.Contains(typeAsString, "DocumentConflictException") {
			return NewDocumentConflictExceptionFromMessage(message)
		}
		return NewConcurrencyException(message)
	}
	// fmt.Printf("ExceptionDispatcher_get2: message='%s', err='%s', typeAsString='%s', code=%d\n", message, err, typeAsString, code)
	// TODO: Java is more complicated, throws exception based on type returned by server.
	// Not sure we can do it in Go
	return NewRavenException("%s", err)
}

func ExceptionDispatcher_throwException(response *http.Response) error {
	if response == nil {
		return NewIllegalArgumentException("Response cannot be null")
	}
	var d []byte
	var err error
	if response.Body != nil {
		d, err = ioutil.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			return NewRavenException("%s", err.Error())
		}
	}
	var schema ExceptionSchema
	if len(d) > 0 {
		err = jsonUnmarshal(d, &schema)
		if response.StatusCode == http.StatusConflict {
			return ExceptionDispatcher_throwConflict(&schema, string(d))
		}
	}

	//fmt.Printf("ExceptionDispatcher_throwException. schema: %#v\n", schema)
	// TODO: Java is more complicated, throws exception based on type returned by server.
	// Not sure we can do it in Go
	return NewRavenException("ExceptionDispatcher_throwException: http response exception")
}

func ExceptionDispatcher_throwConflict(schema *ExceptionSchema, js string) error {
	if strings.Contains(schema.getType(), "DocumentConflictException") {
		return NewDocumentConflictExceptionFromJSON(js)
	}
	return NewConcurrencyException("%s", schema.getMessage())
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
