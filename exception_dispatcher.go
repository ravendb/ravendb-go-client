package ravendb

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
)

func ExceptionDispatcher_get(schema *ExceptionSchema, code int) error {
	// TODO: Java is more complicated
	return ExceptionDispatcher_get2(schema.getMessage(), schema.getError(), schema.getType(), code)
}

func ExceptionDispatcher_get2(message string, err string, typeAsString string, code int) error {
	//fmt.Printf("ExceptionDispatcher_get2: message='%s', err='%s', typeAsString='%s', code=%d\n", message, err, typeAsString, code)
	// TODO: Java is more complicated
	return NewRavenException("ExceptionDispatcher_get: http response exception")
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
		err = json.Unmarshal(d, &schema)
		if response.StatusCode == http.StatusConflict {
			return ExceptionDispatcher_throwConflict(&schema, string(d))
		}
	}

	// TODO: Java is more complicated
	//fmt.Printf("ExceptionDispatcher_throwException. schema: %#v\n", schema)
	//panicIf(true, "More stuff to implement")
	return NewRavenException("ExceptionDispatcher_get: http response exception")
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
