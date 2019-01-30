package ravendb

// PutConnectionStringResult describes result of "put connection" command
type PutConnectionStringResult struct {
	Etag int64 `json:"ETag"`
}
