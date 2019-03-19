package ravendb

import (
	"io/ioutil"
	"net/http"
	"strings"
)

func exceptionDispatcherGetFromSchema(schema *exceptionSchema, code int, inner error) error {
	return exceptionDispatcherGet(schema.Message, schema.Error, schema.Type, code, inner)
}

func exceptionDispatcherGet(message string, errStr string, typeAsString string, code int, inner error) error {
	if code == http.StatusConflict {
		if strings.Contains(typeAsString, "DocumentConflictException") {
			return newDocumentConflictErrorFromMessage(message)
		}
		return newConcurrencyError(message)
	}
	err := exceptionDispatherMakeErrorFromType(typeAsString)
	if err == nil {
		return newRavenError("%s", errStr, inner)
	}

	if errBase, ok := err.(*errorBase); ok {
		errBase.setErrorf("%s", errStr, inner)
	}
	return err
}

func exceptionDispatcherThrowError(response *http.Response) error {
	if response == nil {
		return newIllegalArgumentError("Response cannot be null")
	}
	var d []byte
	var err error
	if response.Body != nil {
		d, err = ioutil.ReadAll(response.Body)
		_ = response.Body.Close()
		if err != nil {
			return newRavenError("%s", err.Error())
		}
	}
	var schema exceptionSchema
	if len(d) > 0 {
		_ = jsonUnmarshal(d, &schema)
		if response.StatusCode == http.StatusConflict {
			return exceptionDispatcherThrowConflict(&schema, string(d))
		}
	}

	err = exceptionDispatherMakeErrorFromType(schema.Type)
	if err == nil {
		return newRavenError("%s. Response: %s", schema.Error, string(d))
	}

	if errBase, ok := err.(*errorBase); ok {
		errBase.setErrorf("%s", schema.Error, err)
	}

	//fmt.Printf("exceptionDispatcherThrowError. schema: %#v\n", schema)
	// TODO: Java is more complicated, throws exception based on type returned by server.
	// Not sure we can do it in Go
	return newRavenError("exceptionDispatcherThrowError: http response exception")
}

func exceptionDispatcherThrowConflict(schema *exceptionSchema, js string) error {
	if strings.Contains(schema.Type, "DocumentConflictException") {
		return newDocumentConflictErrorFromJSON(js)
	}
	return newConcurrencyError("%s", schema.Message)
}

// make an error corresponding to C#'s exception name as returned by the server
func exceptionDispatherMakeErrorFromType(typeAsString string) error {
	if typeAsString == "System.TimeoutException" {
		return &TimeoutError{}
	}

	exceptionName := strings.TrimPrefix(typeAsString, "Raven.Client.Exceptions.")
	if exceptionName == typeAsString {
		return nil
	}
	// those could be further namespaced, take only the last part
	parts := strings.Split(exceptionName, ".")
	if len(parts) > 1 {
		exceptionName = parts[len(parts)-1]
	}
	return makeRavenErrorFromName(typeAsString)
}

type exceptionSchema struct {
	URL     string `json:"Url"`
	Type    string `json:"Type"`
	Message string `json:"Message"`
	Error   string `json:"Error"`
}
