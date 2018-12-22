package ravendb

// DatabaseRecordWithEtag represents database record with etag
type DatabaseRecordWithEtag struct {
	DatabaseRecord
	Etag int `json:"Etag"`
}
