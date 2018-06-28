package ravendb

type AttachmentName struct {
	name        string `json:"Name"`
	hash        string `json:"Hash"`
	contentType string `json:"ContentType"`
	size        int64  `json:"Size"`
}

func (n *AttachmentName) getName() string {
	return n.name
}

func (n *AttachmentName) setName(name string) {
	n.name = name
}

func (n *AttachmentName) getHash() string {
	return n.hash
}
func (n *AttachmentName) setHash(hash string) {
	n.hash = hash
}

func (n *AttachmentName) getContentType() string {
	return n.contentType
}

func (n *AttachmentName) setContentType(contentType string) {
	n.contentType = contentType
}

func (n *AttachmentName) getSize() int64 {
	return n.size
}

func (n *AttachmentName) setSize(size int64) {
	n.size = size
}
