package ravendb

// GetConflictsResult represents result of "get conflict" command
type GetConflictsResult struct {
	ID          string      `json:"Id"`
	Results     []*Conflict `json:"Results"`
	LargestEtag int64       `json:"LargestEtag"`
}

// Conflict represents conflict
type Conflict struct {
	LastModified Time                   `json:"LastModified"`
	ChangeVector string                 `json:"ChangeVector"`
	Doc          map[string]interface{} `json:"Doc"`
}
