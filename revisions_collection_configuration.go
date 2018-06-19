package ravendb

import "time"

type RevisionsCollectionConfiguration struct {
	MinimumRevisionsToKeep   int            `json:"MinimumRevisionsToKeep"`
	MinimumRevisionAgeToKeep *time.Duration `json:"MinimumRevisionAgeToKeep"`
	Disabled                 bool           `json:"Disabled"`
	PurgeOnDelete            bool           `json:"PurgeOnDelete"`
}

func NewRevisionsCollectionConfiguration() *RevisionsCollectionConfiguration {
	return &RevisionsCollectionConfiguration{}
}

func (c *RevisionsCollectionConfiguration) getMinimumRevisionsToKeep() int {
	return c.MinimumRevisionsToKeep
}

func (c *RevisionsCollectionConfiguration) setMinimumRevisionsToKeep(minimumRevisionsToKeep int) {
	c.MinimumRevisionsToKeep = minimumRevisionsToKeep
}

func (c *RevisionsCollectionConfiguration) getMinimumRevisionAgeToKeep() *time.Duration {
	return c.MinimumRevisionAgeToKeep
}

func (c *RevisionsCollectionConfiguration) setMinimumRevisionAgeToKeep(minimumRevisionAgeToKeep *time.Duration) {
	c.MinimumRevisionAgeToKeep = minimumRevisionAgeToKeep
}

func (c *RevisionsCollectionConfiguration) isDisabled() bool {
	return c.Disabled
}

func (c *RevisionsCollectionConfiguration) setDisabled(disabled bool) {
	c.Disabled = disabled
}

func (c *RevisionsCollectionConfiguration) isPurgeOnDelete() bool {
	return c.PurgeOnDelete
}

func (c *RevisionsCollectionConfiguration) setPurgeOnDelete(purgeOnDelete bool) {
	c.PurgeOnDelete = purgeOnDelete
}
