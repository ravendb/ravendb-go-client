package ravendb

// AttachmentDetails represents details of an attachment
type AttachmentDetails struct {
	AttachmentName
	ChangeVector *string `json:"ChangeVector"`
	DocumentID   string  `json:"DocumentId"`
}
