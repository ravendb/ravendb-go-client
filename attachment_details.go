package ravendb

type AttachmentDetails struct {
	AttachmentName
	ChangeVector *string `json:"ChangeVector"`
	DocumentId   string  `json:"DocumentId"`
}

func NewAttachmentDetails() *AttachmentDetails {
	return &AttachmentDetails{}
}

func (d *AttachmentDetails) GetChangeVector() *string {
	return d.ChangeVector
}

func (d *AttachmentDetails) SetChangeVector(changeVector *string) {
	d.ChangeVector = changeVector
}

func (d *AttachmentDetails) GetDocumentID() string {
	return d.DocumentId
}

func (d *AttachmentDetails) SetDocumentID(documentId string) {
	d.DocumentId = documentId
}
