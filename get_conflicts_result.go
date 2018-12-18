package ravendb

type GetConflictsResult struct {
	ID          string      `json:"Id"`
	Results     []*Conflict `json:"Results"`
	LargestEtag int         `json:"LargestEtag"`
}

type Conflict struct {
	LastModified Time       `json:"LastModified"`
	ChangeVector string     `json:"ChangeVector"`
	Doc          ObjectNode `json:"Doc"`
}
