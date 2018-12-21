package ravendb

// AttachmentName represents infor about an attachment
type AttachmentName struct {
	Name        string `json:"Name"`
	Hash        string `json:"Hash"`
	ContentType string `json:"ContentType"`
	Size        int64  `json:"Size"`
}
