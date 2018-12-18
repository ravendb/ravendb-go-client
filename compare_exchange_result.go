package ravendb

import (
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
		exchangeResult := NewCompareExchangeResult()
		exchangeResult.index = index
		exchangeResult.value = Defaults_defaultValue(clazz)
		exchangeResult.successful = successful
		return exchangeResult, nil
	}

	result, err := convertValue(val, clazz)
	if err != nil {
		return nil, err
	}

	exchangeResult := NewCompareExchangeResult()
	exchangeResult.index = index
	exchangeResult.value = result
	exchangeResult.successful = successful
	return exchangeResult, nil
}

func (r *CompareExchangeResult) GetValue() interface{} {
	return r.value
}

func (r *CompareExchangeResult) SetValue(value interface{}) {
	r.value = value
}

func (r *CompareExchangeResult) GetIndex() int {
	return r.index
}

func (r *CompareExchangeResult) SetIndex(index int) {
	r.index = index
}

func (r *CompareExchangeResult) IsSuccessful() bool {
	return r.successful
}

func (r *CompareExchangeResult) SetSuccessful(successful bool) {
	r.successful = successful
}
