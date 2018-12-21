package ravendb

// PutIndexResult represents result of put index command
type PutIndexResult struct {
	// Note: don't know how Java does it, but this is sent
	// as Index in JSON responses from the server
	IndexName string `json:"Index"`
}
