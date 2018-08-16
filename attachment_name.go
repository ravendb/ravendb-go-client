package ravendb

type AttachmentName struct {
	Name        string `json:"Name"`
	Hash        string `json:"Hash"`
	ContentType string `json:"ContentType"`
	Size        int64  `json:"Size"`
}

func (n *AttachmentName) GetName() string {
	return n.Name
}

func (n *AttachmentName) SetName(name string) {
	n.Name = name
}

func (n *AttachmentName) GetHash() string {
	return n.Hash
}
func (n *AttachmentName) SetHash(hash string) {
	n.Hash = hash
}

func (n *AttachmentName) GetContentType() string {
	return n.ContentType
}

func (n *AttachmentName) SetContentType(contentType string) {
	n.ContentType = contentType
}

func (n *AttachmentName) GetSize() int64 {
	return n.Size
}

func (n *AttachmentName) SetSize(size int64) {
	n.Size = size
}
