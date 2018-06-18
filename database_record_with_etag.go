package ravendb

type DatabaseRecordWithEtag struct {
	// TODO: should I just replicate the fields? Not sure how json deals with this
	DatabaseRecord

	Etag int `json:"Etag"`
}

func (r *DatabaseRecordWithEtag) getEtag() int {
	return r.Etag
}

func (r *DatabaseRecordWithEtag) setEtag(etag int) {
	r.Etag = etag
}
