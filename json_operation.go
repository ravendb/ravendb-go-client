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

func JsonOperation_compareJson(id String, originalJson ObjectNode, newJson ObjectNode, changes map[string][]*DocumentsChanges, docChanges *[]*DocumentsChanges) bool {
	panicIf(true, "NYI")
	return false
}

func JsonOperation_newChange(name String, newValue Object, oldValue Object, docChanges *[]*DocumentsChanges, change ChangeType) {
	/*
		TODO:
		if (newValue instanceof NumericNode) {
			NumericNode node = (NumericNode) newValue;
			newValue = node.numberValue();
		}

		if (oldValue instanceof NumericNode) {
			NumericNode node = (NumericNode) oldValue;
			oldValue = node.numberValue();
		}
	*/

	documentsChanges := NewDocumentsChanges()
	documentsChanges.setFieldName(name)
	documentsChanges.setFieldNewValue(newValue)
	documentsChanges.setFieldOldValue(oldValue)
	documentsChanges.setChange(change)
	*docChanges = append(*docChanges, documentsChanges)
}
