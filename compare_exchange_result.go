package ravendb

import (
	"reflect"
)

// CompareExchangeResult describes result of compare exchange
type CompareExchangeResult struct {
	Value        interface{}
	Index        int
	IsSuccessful bool
}

func parseCompareExchangeResultFromString(clazz reflect.Type, responseString []byte, conventions *DocumentConventions) (*CompareExchangeResult, error) {
	var response map[string]interface{}
	err := jsonUnmarshal(responseString, &response)
	if err != nil {
		return nil, err
	}
	index, ok := jsonGetAsInt(response, "Index")
	if !ok {
		return nil, newIllegalStateError("Response is invalid. Index is missing")
	}

	successful, _ := jsonGetAsBool(response, "Successful")

	var val interface{}
	raw, ok := response["Value"]
	if ok && raw != nil {
		if m, ok := raw.(map[string]interface{}); ok {
			val = m["Object"]
		}
	}

	if val == nil {
		exchangeResult := &CompareExchangeResult{}
		exchangeResult.Index = index
		exchangeResult.Value = getDefaultValueForType(clazz)
		exchangeResult.IsSuccessful = successful
		return exchangeResult, nil
	}

	result, err := convertValue(val, clazz)
	if err != nil {
		return nil, err
	}

	exchangeResult := &CompareExchangeResult{}
	exchangeResult.Index = index
	exchangeResult.Value = result
	exchangeResult.IsSuccessful = successful
	return exchangeResult, nil
}
