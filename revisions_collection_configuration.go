package ravendb

// RevisionsCollectionConfiguration describes configuration for revisions collection
type RevisionsCollectionConfiguration struct {
	// Note: in java MinimumRevisionsToKeep is Long, which is ref type
	// so by default it's not serialized to JSON if not set
	// in Go 0 is default so use "omitempty" to achieve the same effect
	MinimumRevisionsToKeep   int64     `json:"MinimumRevisionsToKeep,omitempty"`
	MinimumRevisionAgeToKeep *Duration `json:"MinimumRevisionAgeToKeep"`
	Disabled                 bool      `json:"Disabled"`
	PurgeOnDelete            bool      `json:"PurgeOnDelete"`
}
