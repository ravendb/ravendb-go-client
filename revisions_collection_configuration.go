package ravendb

import "time"

type RevisionsCollectionConfiguration struct {
	MinimumRevisionsToKeep   int            `json:"MinimumRevisionsToKeep"`
	MinimumRevisionAgeToKeep *time.Duration `json:"MinimumRevisionAgeToKeep"`
	Disabled                 bool           `json:"Disabled"`
	PurgeOnDelete            bool           `json:"PurgeOnDelete"`
}
