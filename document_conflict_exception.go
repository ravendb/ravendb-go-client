package ravendb

type DocumentConflictException struct {
	*ConflictException
	DocID       string
	LargestEtag int
}

func NewDocumentConflictException(message string, docID string, etag int) *DocumentConflictException {
	res := &DocumentConflictException{}
	res.ConflictException = NewConflictException("%s", message)
	res.DocID = docID
	res.LargestEtag = etag
	return res
}

func NewDocumentConflictExceptionFromMessage(message string) *DocumentConflictException {
	return NewDocumentConflictException(message, "", 0)
}

func NewDocumentConflictExceptionFromJSON(js string) error {
	var jsonNode map[string]interface{}
	err := jsonUnmarshal([]byte(js), &jsonNode)
	if err != nil {
		return NewBadResponseException("Unable to parse server response: %s", err)
	}
	docID, _ := JsonGetAsText(jsonNode, "DocId")
	message, _ := JsonGetAsText(jsonNode, "Message")
	largestEtag, _ := jsonGetAsInt(jsonNode, "LargestEtag")

	return NewDocumentConflictException(message, docID, largestEtag)
}
