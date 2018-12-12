package ravendb

import (
	"fmt"
	"sort"
)

func JsonOperation_entityChanged(newObj ObjectNode, documentInfo *DocumentInfo, changes map[string][]*DocumentsChanges) bool {
	var docChanges []*DocumentsChanges

	doc := documentInfo.document
	if !documentInfo.newDocument && doc != nil {
		id := documentInfo.id
		return JsonOperation_compareJson(id, doc, newObj, changes, &docChanges)
	}

	if changes == nil {
		return true
	}

	JsonOperation_newChange("", nil, nil, &docChanges, DocumentsChanges_ChangeType_DOCUMENT_ADDED)
	id := documentInfo.id
	a := changes[id]
	a = append(a, docChanges...)
	changes[id] = a
	return true
}

func isJSONFloatEqual(oldPropVal float64, newProp interface{}) bool {
	switch newPropVal := newProp.(type) {
	case float64:
		return oldPropVal == newPropVal
	default:
		// TODO: can those happen in real life?
		panicIf(true, "unhandled type of newProp, expected 'float64' and is '%T'", newProp)
	}
	return false
}

func isJSONBoolEqual(oldPropVal bool, newProp interface{}) bool {
	switch newPropVal := newProp.(type) {
	case bool:
		return oldPropVal == newPropVal
	default:
		// TODO: can those happen in real life?
		panicIf(true, "unhandled type of newProp, expected 'bool' and is '%T'", newProp)
	}
	return false
}

func isJSONStringEqual(oldPropVal string, newProp interface{}) bool {
	switch newPropVal := newProp.(type) {
	case string:
		return oldPropVal == newPropVal
	default:
		// TODO: can those happen in real life?
		panicIf(true, "unhandled type of newProp, expected 'string' and is '%T'", newProp)
	}
	return false
}

func JsonOperation_compareJson(id string, originalJson ObjectNode, newJson ObjectNode, changes map[string][]*DocumentsChanges, docChanges *[]*DocumentsChanges) bool {
	newJsonProps := getObjectNodeFieldNames(newJson)
	oldJsonProps := getObjectNodeFieldNames(originalJson)
	newFields := StringArraySubtract(newJsonProps, oldJsonProps)
	removedFields := StringArraySubtract(oldJsonProps, newJsonProps)

	for _, field := range removedFields {
		if changes == nil {
			return true
		}
		JsonOperation_newChange(field, nil, nil, docChanges, DocumentsChanges_ChangeType_REMOVED_FIELD)
	}

	for _, prop := range newJsonProps {
		switch prop {
		case Constants_Documents_Metadata_LAST_MODIFIED,
			Constants_Documents_Metadata_COLLECTION,
			Constants_Documents_Metadata_CHANGE_VECTOR,
			Constants_Documents_Metadata_ID:
			continue
		}
		if StringArrayContains(newFields, prop) {
			if changes == nil {
				return true
			}
			v := newJson[prop]
			JsonOperation_newChange(prop, v, nil, docChanges, DocumentsChanges_ChangeType_NEW_FIELD)
			continue
		}
		newProp := newJson[prop]
		oldProp := originalJson[prop]
		switch newPropVal := newProp.(type) {
		case float64:
			if isJSONFloatEqual(newPropVal, oldProp) {
				break
			}
			if changes == nil {
				return true
			}
			JsonOperation_newChange(prop, newProp, oldProp, docChanges, DocumentsChanges_ChangeType_FIELD_CHANGED)
		case string:
			if isJSONStringEqual(newPropVal, oldProp) {
				break
			}
			if changes == nil {
				return true
			}
			JsonOperation_newChange(prop, newProp, oldProp, docChanges, DocumentsChanges_ChangeType_FIELD_CHANGED)
		case bool:
			isJSONBoolEqual(newPropVal, oldProp)
		case []interface{}:
			if oldProp == nil || !isInstanceOfArrayOfInterface(oldProp) {
				if changes == nil {
					return true
				}

				JsonOperation_newChange(prop, newProp, oldProp, docChanges, DocumentsChanges_ChangeType_FIELD_CHANGED)
				break
			}

			changed := JsonOperation_compareJsonArray(id, oldProp.([]interface{}), newProp.([]interface{}), changes, docChanges, prop)
			if changes == nil && changed {
				return true
			}

		case map[string]interface{}:
			oldPropVal, ok := oldProp.(map[string]interface{})
			// TODO: a better check for nil?
			if !ok || oldProp == nil {
				if changes == nil {
					return true
				}
				JsonOperation_newChange(prop, newProp, nil, docChanges, DocumentsChanges_ChangeType_FIELD_CHANGED)
				break
			}
			changed := JsonOperation_compareJson(id, oldPropVal, newPropVal, changes, docChanges)
			if changes == nil && changed {
				return true
			}
		default:
			if newProp == nil {
				if oldProp == nil {
					break
				}
				if changes == nil {
					return true
				}
				JsonOperation_newChange(prop, nil, oldProp, docChanges, DocumentsChanges_ChangeType_FIELD_CHANGED)
				break
			}
			// TODO: array, nil
			// Write tests for all types
			panicIf(true, "unhandled type %T, newProp: '%v', oldProp: '%v'", newProp, newProp, oldProp)
		}
	}

	if changes == nil || len(*docChanges) == 0 {
		return false
	}
	changes[id] = *docChanges
	return true
}

func isInstanceOfArrayOfInterface(v interface{}) bool {
	_, ok := v.([]interface{})
	return ok
}

func JsonOperation_compareJsonArray(id string, oldArray []interface{}, newArray []interface{}, changes map[string][]*DocumentsChanges, docChanges *[]*DocumentsChanges, propName string) bool {
	// if we don't care about the changes
	if len(oldArray) != len(newArray) && changes == nil {
		return true
	}

	changed := false

	position := 0
	maxPos := len(oldArray)
	if maxPos > len(newArray) {
		maxPos = len(newArray)
	}
	for ; position < maxPos; position++ {
		oldVal := oldArray[position]
		newVal := newArray[position]
		switch oldVal.(type) {
		case ObjectNode:
			if _, ok := newVal.(ObjectNode); ok {
				newChanged := JsonOperation_compareJson(id, oldVal.(ObjectNode), newVal.(ObjectNode), changes, docChanges)
				if newChanged {
					changed = newChanged
				}
			} else {
				changed = true
				if changes != nil {
					JsonOperation_newChange(propName, newVal, oldVal, docChanges, DocumentsChanges_ChangeType_ARRAY_VALUE_ADDED)
				}
			}
		case []interface{}:
			if _, ok := newVal.([]interface{}); ok {
				newChanged := JsonOperation_compareJsonArray(id, oldVal.([]interface{}), newVal.([]interface{}), changes, docChanges, propName)
				if newChanged {
					changed = newChanged
				}
			} else {
				changed = true
				if changes != nil {
					JsonOperation_newChange(propName, newVal, oldVal, docChanges, DocumentsChanges_ChangeType_ARRAY_VALUE_CHANGED)
				}
			}
		default:
			// NULL case
			if oldVal == nil {
				if newVal != nil {
					changed = true
					if changes != nil {
						JsonOperation_newChange(propName, newVal, oldVal, docChanges, DocumentsChanges_ChangeType_ARRAY_VALUE_CHANGED)
					}
				}
				break
			}
			// Note: this matches Java but also means that 1 == "1"
			oldValStr := fmt.Sprintf("%v", oldVal)
			newValStr := fmt.Sprintf("%v", newVal)
			if oldValStr != newValStr {
				if changes != nil {
					JsonOperation_newChange(propName, newVal, oldVal, docChanges, DocumentsChanges_ChangeType_ARRAY_VALUE_CHANGED)
				}
				changed = true
			}

		}
	}

	if changes == nil {
		return changed
	}

	// if one of the arrays is larger than the other
	for ; position < len(oldArray); position++ {
		JsonOperation_newChange(propName, nil, oldArray[position], docChanges, DocumentsChanges_ChangeType_ARRAY_VALUE_REMOVED)
	}

	for ; position < len(newArray); position++ {
		JsonOperation_newChange(propName, newArray[position], nil, docChanges, DocumentsChanges_ChangeType_ARRAY_VALUE_ADDED)
	}

	return changed
}

func JsonOperation_newChange(name string, newValue interface{}, oldValue interface{}, docChanges *[]*DocumentsChanges, change ChangeType) {
	documentsChanges := NewDocumentsChanges()
	documentsChanges.setFieldName(name)
	documentsChanges.setFieldNewValue(newValue)
	documentsChanges.setFieldOldValue(oldValue)
	documentsChanges.setChange(change)
	*docChanges = append(*docChanges, documentsChanges)
}

func getObjectNodeFieldNames(o ObjectNode) []string {
	n := len(o)
	if n == 0 {
		return nil
	}
	res := make([]string, n, n)
	i := 0
	for k := range o {
		res[i] = k
		i++
	}
	// Go randomizes order of map tranversal but it's useful to have it
	// fixed e.g. for tests
	sort.Strings(res)
	return res
}
