package ravendb

import (
	"encoding/json"
	"reflect"
)

type CompareExchangeResult struct {
	value      interface{}
	index      int
	successful bool
}

func NewCompareExchangeResult() *CompareExchangeResult {
	return &CompareExchangeResult{}
}

func CompareExchangeResult_parseFromString(clazz reflect.Type, responseString []byte, conventions *DocumentConventions) (*CompareExchangeResult, error) {
	var response map[string]interface{}
	err := json.Unmarshal(responseString, &response)
	if err != nil {
		return nil, err
	}
	index, ok := jsonGetAsInt(response, "Index")
	if !ok {
		return nil, NewIllegalStateException("Response is invalid. Index is missing")
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
		exchangeResult := NewCompareExchangeResult()
		exchangeResult.index = index
		exchangeResult.value = Defaults_defaultValue(clazz)
		exchangeResult.successful = successful
		return exchangeResult, nil
	}

	result, err := convertValue(val, clazz)

	exchangeResult := NewCompareExchangeResult()
	exchangeResult.index = index
	exchangeResult.value = result
	exchangeResult.successful = successful
	return exchangeResult, nil
}

func (r *CompareExchangeResult) getValue() interface{} {
	return r.value
}

func (r *CompareExchangeResult) setValue(value interface{}) {
	r.value = value
}

func (r *CompareExchangeResult) getIndex() int {
	return r.index
}

func (r *CompareExchangeResult) setIndex(index int) {
	r.index = index
}

func (r *CompareExchangeResult) isSuccessful() bool {
	return r.successful
}

func (r *CompareExchangeResult) setSuccessful(successful bool) {
	r.successful = successful
}
