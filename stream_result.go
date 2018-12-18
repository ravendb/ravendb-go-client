package ravendb

type StreamResult struct {
	ID           string
	changeVector *string
	metadata     *MetadataAsDictionary
	document     interface{}
}
