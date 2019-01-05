package ravendb

import (
	"reflect"
)

func compareExchangeValueResultParserGetValues(clazz reflect.Type, response []byte, conventions *DocumentConventions) (map[string]*CompareExchangeValue, error) {
	var jsonResponse map[string]interface{}
	err := jsonUnmarshal(response, &jsonResponse)
	if err != nil {
		return nil, err
	}

	results := make(map[string]*CompareExchangeValue)

	itemsI, ok := jsonResponse["Results"]
	if !ok {
		return nil, newIllegalStateError("Response is invalid. Results is missing.")
	}
	items, ok := itemsI.([]interface{})
	if !ok {
		return nil, newIllegalStateError("Response is invalid. Results is missing.")
	}

	for _, itemI := range items {
		if itemI == nil {
			return nil, newIllegalStateError("Response is invalid. Item is null")
		}
		item, ok := itemI.(map[string]interface{})
		if !ok {
			return nil, newIllegalStateError("Response is invalid. Item is null")
		}
		key, ok := jsonGetAsString(item, "Key")
		if !ok {
			return nil, newIllegalStateError("Response is invalid. Key is missing.")
		}

		index, ok := jsonGetAsInt(item, "Index")

		if !ok {
			return nil, newIllegalStateError("Response is invalid. Index is missing")
		}

		raw, ok := item["Value"]
		if !ok || raw == nil {
			return nil, newIllegalStateError("Response is invalid. Value is missing.")
		}
		rawMap, ok := raw.(map[string]interface{})
		if !ok {
			return nil, newIllegalStateError("Response is invalid. Value is missing.")
		}

		if isTypePrimitive(clazz) {
			var value interface{}
			rawValue := rawMap["Object"]
			value, err = convertValue(rawValue, clazz)
			if err != nil {
				return nil, err
			}
			cmpValue := NewCompareExchangeValue(key, index, value)
			results[key] = cmpValue
		} else {
			object, ok := rawMap["Object"]
			if !ok || object == nil {
				v := NewCompareExchangeValue(key, index, getDefaultValueForType(clazz))
				results[key] = v
			} else {
				converted, err := convertValue(object, clazz)
				if err != nil {
					return nil, err
				}
				v := NewCompareExchangeValue(key, index, converted)
				results[key] = v
			}
		}
	}

	return results, nil
}

func compareExchangeValueResultParserGetValue(clazz reflect.Type, response []byte, conventions *DocumentConventions) (*CompareExchangeValue, error) {
	if response == nil {
		return nil, nil
	}

	values, err := compareExchangeValueResultParserGetValues(clazz, response, conventions)
	if err != nil {
		return nil, err
	}
	if len(values) == 0 {
		return nil, nil
	}
	// Note: in Go iteration order is random, so this might behave
	// differently than Java
	for _, v := range values {
		return v, nil
	}
	panicIf(true, "Should never be reached")
	return nil, nil
}
