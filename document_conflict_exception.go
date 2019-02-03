package ravendb

// DocumentConflictError represents document conflict error from the server
type DocumentConflictError struct {
	*ConflictException
	DocID       string
	LargestEtag int64
}

func newDocumentConflictError(message string, docID string, etag int64) *DocumentConflictError {
	res := &DocumentConflictError{}
	res.ConflictException = NewConflictException("%s", message)
	res.DocID = docID
	res.LargestEtag = etag
	return res
}

func newDocumentConflictErrorFromMessage(message string) *DocumentConflictError {
	return newDocumentConflictError(message, "", 0)
}

func newDocumentConflictErrorFromJSON(js string) error {
	var jsonNode map[string]interface{}
	err := jsonUnmarshal([]byte(js), &jsonNode)
	if err != nil {
		return newBadResponseError("Unable to parse server response: %s", err)
	}
	docID, _ := jsonGetAsText(jsonNode, "DocId")
	message, _ := jsonGetAsText(jsonNode, "Message")
	largestEtag, _ := jsonGetAsInt64(jsonNode, "LargestEtag")

	return newDocumentConflictError(message, docID, largestEtag)
}
