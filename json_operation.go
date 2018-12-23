package ravendb

import (
	"fmt"
	"sort"
)

func jsonOperationEntityChanged(newObj ObjectNode, documentInfo *documentInfo, changes map[string][]*DocumentsChanges) bool {
	var docChanges []*DocumentsChanges

	doc := documentInfo.document
	if !documentInfo.newDocument && doc != nil {
		id := documentInfo.id
		return jsonOperationCompareJson(id, doc, newObj, changes, &docChanges)
	}

	if changes == nil {
		return true
	}

	jsonOperationNewChange("", nil, nil, &docChanges, DocumentChangeDocumentAdded)
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

func jsonOperationCompareJson(id string, originalJson ObjectNode, newJson ObjectNode, changes map[string][]*DocumentsChanges, docChanges *[]*DocumentsChanges) bool {
	newJsonProps := getObjectNodeFieldNames(newJson)
	oldJsonProps := getObjectNodeFieldNames(originalJson)
	newFields := stringArraySubtract(newJsonProps, oldJsonProps)
	removedFields := stringArraySubtract(oldJsonProps, newJsonProps)

	for _, field := range removedFields {
		if changes == nil {
			return true
		}
		jsonOperationNewChange(field, nil, nil, docChanges, DocumentChangeRemovedField)
	}

	for _, prop := range newJsonProps {
		switch prop {
		case MetadataLastModified,
			MetadataCollection,
			MetadataChangeVector,
			MetadataID:
			continue
		}
		if stringArrayContains(newFields, prop) {
			if changes == nil {
				return true
			}
			v := newJson[prop]
			jsonOperationNewChange(prop, v, nil, docChanges, DocumentChangeNewField)
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
			jsonOperationNewChange(prop, newProp, oldProp, docChanges, DocumentChangeFieldChanged)
		case string:
			if isJSONStringEqual(newPropVal, oldProp) {
				break
			}
			if changes == nil {
				return true
			}
			jsonOperationNewChange(prop, newProp, oldProp, docChanges, DocumentChangeFieldChanged)
		case bool:
			isJSONBoolEqual(newPropVal, oldProp)
		case []interface{}:
			if oldProp == nil || !isInstanceOfArrayOfInterface(oldProp) {
				if changes == nil {
					return true
				}

				jsonOperationNewChange(prop, newProp, oldProp, docChanges, DocumentChangeFieldChanged)
				break
			}

			changed := jsonOperationCompareJsonArray(id, oldProp.([]interface{}), newProp.([]interface{}), changes, docChanges, prop)
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
				jsonOperationNewChange(prop, newProp, nil, docChanges, DocumentChangeFieldChanged)
				break
			}
			changed := jsonOperationCompareJson(id, oldPropVal, newPropVal, changes, docChanges)
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
				jsonOperationNewChange(prop, nil, oldProp, docChanges, DocumentChangeFieldChanged)
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

func jsonOperationCompareJsonArray(id string, oldArray []interface{}, newArray []interface{}, changes map[string][]*DocumentsChanges, docChanges *[]*DocumentsChanges, propName string) bool {
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
				newChanged := jsonOperationCompareJson(id, oldVal.(ObjectNode), newVal.(ObjectNode), changes, docChanges)
				if newChanged {
					changed = newChanged
				}
			} else {
				changed = true
				if changes != nil {
					jsonOperationNewChange(propName, newVal, oldVal, docChanges, DocumentChangeArrayValueAdded)
				}
			}
		case []interface{}:
			if _, ok := newVal.([]interface{}); ok {
				newChanged := jsonOperationCompareJsonArray(id, oldVal.([]interface{}), newVal.([]interface{}), changes, docChanges, propName)
				if newChanged {
					changed = newChanged
				}
			} else {
				changed = true
				if changes != nil {
					jsonOperationNewChange(propName, newVal, oldVal, docChanges, DocumentChangeArrayValueChanged)
				}
			}
		default:
			// NULL case
			if oldVal == nil {
				if newVal != nil {
					changed = true
					if changes != nil {
						jsonOperationNewChange(propName, newVal, oldVal, docChanges, DocumentChangeArrayValueChanged)
					}
				}
				break
			}
			// Note: this matches Java but also means that 1 == "1"
			oldValStr := fmt.Sprintf("%v", oldVal)
			newValStr := fmt.Sprintf("%v", newVal)
			if oldValStr != newValStr {
				if changes != nil {
					jsonOperationNewChange(propName, newVal, oldVal, docChanges, DocumentChangeArrayValueChanged)
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
		jsonOperationNewChange(propName, nil, oldArray[position], docChanges, DocumentChangeArrayValueRemoved)
	}

	for ; position < len(newArray); position++ {
		jsonOperationNewChange(propName, newArray[position], nil, docChanges, DocumentChangeArrayValueAdded)
	}

	return changed
}

func jsonOperationNewChange(name string, newValue interface{}, oldValue interface{}, docChanges *[]*DocumentsChanges, change ChangeType) {
	documentsChanges := &DocumentsChanges{
		FieldNewValue: newValue,
		FieldOldValue: oldValue,
		FieldName:     name,
		Change:        change,
	}
	*docChanges = append(*docChanges, documentsChanges)
}

func getObjectNodeFieldNames(o ObjectNode) []string {
	n := len(o)
	if n == 0 {
		return nil
	}
	res := make([]string, n)
	i := 0
	for k := range o {
		res[i] = k
		i++
	}
	// Go randomizes order of map traversal but it's useful to have it
	// fixed e.g. for tests
	sort.Strings(res)
	return res
}
