package ravendb

type StreamResult struct {
	ID           string
	changeVector *string
	metadata     *IMetadataDictionary
	document     interface{}
}
