package ravendb

type AttachmentDetails struct {
	AttachmentName
	ChangeVector *string `json:"ChangeVector"`
	DocumentID   string  `json:"DocumentId"`
}
