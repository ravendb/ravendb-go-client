package ravendb

// DocumentConflictError represents document conflict error from the server
type DocumentConflictError struct {
	*ConflictException
	DocID       string
	LargestEtag int
}

func NewDocumentConflictError(message string, docID string, etag int) *DocumentConflictError {
	res := &DocumentConflictError{}
	res.ConflictException = NewConflictException("%s", message)
	res.DocID = docID
	res.LargestEtag = etag
	return res
}

func NewDocumentConflictErrorFromMessage(message string) *DocumentConflictError {
	return NewDocumentConflictError(message, "", 0)
}

func NewDocumentConflictErrorFromJSON(js string) error {
	var jsonNode map[string]interface{}
	err := jsonUnmarshal([]byte(js), &jsonNode)
	if err != nil {
		return newBadResponseError("Unable to parse server response: %s", err)
	}
	docID, _ := JsonGetAsText(jsonNode, "DocId")
	message, _ := JsonGetAsText(jsonNode, "Message")
	largestEtag, _ := jsonGetAsInt(jsonNode, "LargestEtag")

	return NewDocumentConflictError(message, docID, largestEtag)
}
