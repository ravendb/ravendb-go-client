package ravendb

import "encoding/json"

type DocumentConflictException struct {
	*ConflictException
	DocId       string
	LargestEtag int
}

func NewDocumentConflictException(message string, docId string, etag int) *DocumentConflictException {
	res := &DocumentConflictException{}
	res.ConflictException = NewConflictException("%s", message)
	res.DocId = docId
	res.LargestEtag = etag
	return res
}

func NewDocumentConflictExceptionFromMessage(message string) *DocumentConflictException {
	return NewDocumentConflictException(message, "", 0)
}

func NewDocumentConflictExceptionFromJSON(js string) error {
	var jsonNode map[string]interface{}
	err := json.Unmarshal([]byte(js), &jsonNode)
	if err != nil {
		return NewBadResponseException("Unable to parse server response: %s", err)
	}
	docID, _ := jsonGetAsText(jsonNode, "DocId")
	message, _ := jsonGetAsText(jsonNode, "Message")
	largestEtag, _ := jsonGetAsInt(jsonNode, "LargestEtag")

	return NewDocumentConflictException(message, docID, largestEtag)
}

/*
   public String getDocId() {
       return docId;
   }

   public void setDocId(String docId) {
       this.docId = docId;
   }

   public long getLargestEtag() {
       return largestEtag;
   }

   public void setLargestEtag(long largestEtag) {
       this.largestEtag = largestEtag;
   }
*/
