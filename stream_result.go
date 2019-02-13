package ravendb

// StreamResult represents result of stream iterator
type StreamResult struct {
	ID           string
	ChangeVector *string
	Metadata     *MetadataAsDictionary
	Document     interface{}
}
