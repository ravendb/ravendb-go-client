package ravendb

import (
	"io/ioutil"
	"net/http"
	"strings"
)

func exceptionDispatcherGet(schema *ExceptionSchema, code int) error {
	return exceptionDispatcherGet2(schema.Message, schema.Error, schema.Type, code)
}

func exceptionDispatcherGet2(message string, err string, typeAsString string, code int) error {
	if code == http.StatusConflict {
		if strings.Contains(typeAsString, "DocumentConflictError") {
			return newDocumentConflictErrorFromMessage(message)
		}
		return newConcurrencyError(message)
	}
	// fmt.Printf("exceptionDispatcherGet2: message='%s', err='%s', typeAsString='%s', code=%d\n", message, err, typeAsString, code)
	// TODO: Java is more complicated, throws exception based on type returned by server.
	// Not sure we can do it in Go
	return newRavenError("%s", err)
}

func exceptionDispatcherThrowError(response *http.Response) error {
	if response == nil {
		return newIllegalArgumentError("Response cannot be null")
	}
	var d []byte
	var err error
	if response.Body != nil {
		d, err = ioutil.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			return newRavenError("%s", err.Error())
		}
	}
	var schema ExceptionSchema
	if len(d) > 0 {
		jsonUnmarshal(d, &schema)
		if response.StatusCode == http.StatusConflict {
			return exceptionDispatcherThrowConflict(&schema, string(d))
		}
	}

	//fmt.Printf("exceptionDispatcherThrowError. schema: %#v\n", schema)
	// TODO: Java is more complicated, throws exception based on type returned by server.
	// Not sure we can do it in Go
	return newRavenError("exceptionDispatcherThrowError: http response exception")
}

func exceptionDispatcherThrowConflict(schema *ExceptionSchema, js string) error {
	if strings.Contains(schema.Type, "DocumentConflictError") {
		return newDocumentConflictErrorFromJSON(js)
	}
	return newConcurrencyError("%s", schema.Message)
}

type ExceptionSchema struct {
	URL     string `json:"Url"`
	Type    string `json:"Type"`
	Message string `json:"Message"`
	Error   string `json:"Error"`
}
