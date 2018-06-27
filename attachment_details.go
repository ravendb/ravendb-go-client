package ravendb

type AttachmentDetails struct {
	AttachmentName
	changeVector string `json:"ChangeVector"`
	documentId   string `json:"DocumentId"`
}

func (d *AttachmentDetails) getChangeVector() string {
	return d.changeVector
}

func (d *AttachmentDetails) setChangeVector(changeVector string) {
	d.changeVector = changeVector
}

func (d *AttachmentDetails) getDocumentId() string {
	return d.documentId
}

func (d *AttachmentDetails) setDocumentId(documentId string) {
	d.documentId = documentId
}
