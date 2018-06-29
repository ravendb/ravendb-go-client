package ravendb

type AttachmentDetails struct {
	AttachmentName
	ChangeVector *string `json:"ChangeVector"`
	DocumentId   string  `json:"DocumentId"`
}

func NewAttachmentDetails() *AttachmentDetails {
	return &AttachmentDetails{}
}

func (d *AttachmentDetails) getChangeVector() *string {
	return d.ChangeVector
}

func (d *AttachmentDetails) setChangeVector(changeVector *string) {
	d.ChangeVector = changeVector
}

func (d *AttachmentDetails) getDocumentId() string {
	return d.DocumentId
}

func (d *AttachmentDetails) setDocumentId(documentId string) {
	d.DocumentId = documentId
}
