package ravendb

type AttachmentName struct {
	Name        string `json:"Name"`
	Hash        string `json:"Hash"`
	ContentType string `json:"ContentType"`
	Size        int64  `json:"Size"`
}

func (n *AttachmentName) getName() string {
	return n.Name
}

func (n *AttachmentName) setName(name string) {
	n.Name = name
}

func (n *AttachmentName) getHash() string {
	return n.Hash
}
func (n *AttachmentName) setHash(hash string) {
	n.Hash = hash
}

func (n *AttachmentName) getContentType() string {
	return n.ContentType
}

func (n *AttachmentName) setContentType(contentType string) {
	n.ContentType = contentType
}

func (n *AttachmentName) getSize() int64 {
	return n.Size
}

func (n *AttachmentName) setSize(size int64) {
	n.Size = size
}
