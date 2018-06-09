package ravendb

func JsonOperation_entityChanged(newObj ObjectNode, documentInfo *DocumentInfo, changes map[string][]*DocumentsChanges) bool {
	var docChanges []*DocumentsChanges

	doc := documentInfo.getDocument()
	if !documentInfo.isNewDocument() && doc != nil {
		id := documentInfo.getId()
		return JsonOperation_compareJson(id, doc, newObj, changes, &docChanges)
	}

	if changes == nil {
		return true
	}

	JsonOperation_newChange("", nil, nil, &docChanges, DocumentsChanges_ChangeType_DOCUMENT_ADDED)
	id := documentInfo.getId()
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
		// newProp should be
		panicIf(true, "unhandled type")
	}
	return false
}

func isJSONStringEqual(oldPropVal string, newProp interface{}) bool {
	switch newPropVal := newProp.(type) {
	case string:
		return oldPropVal == newPropVal
	default:
		// TODO: can those happen in real life?
		// newProp should be
		panicIf(true, "unhandled type")
	}
	return false
}

func JsonOperation_compareJson(id string, originalJson ObjectNode, newJson ObjectNode, changes map[string][]*DocumentsChanges, docChanges *[]*DocumentsChanges) bool {
	newJsonProps := getObjectNodeFieldNames(newJson)
	oldJsonProps := getObjectNodeFieldNames(originalJson)
	newFields := stringArraySubtract(newJsonProps, oldJsonProps)
	removedFields := stringArraySubtract(oldJsonProps, newJsonProps)

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
		if stringArrayContains(newFields, prop) {
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
			panicIf(true, "unhandled type %T, newProp: '%v', oldProp: '%s'", newProp, newProp, oldProp)
		}
	}

	if changes == nil || len(*docChanges) == 0 {
		return false
	}
	changes[id] = *docChanges
	return true
}

func JsonOperation_newChange(name string, newValue Object, oldValue Object, docChanges *[]*DocumentsChanges, change ChangeType) {
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
	return res
}
