package ravendb

// DatabaseRecordWithEtag represents database record with etag
type DatabaseRecordWithEtag struct {
	DatabaseRecord
	Etag int64 `json:"Etag"`
}
