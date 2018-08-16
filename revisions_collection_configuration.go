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

func (c *RevisionsCollectionConfiguration) GetMinimumRevisionsToKeep() int {
	return c.MinimumRevisionsToKeep
}

func (c *RevisionsCollectionConfiguration) SetMinimumRevisionsToKeep(minimumRevisionsToKeep int) {
	c.MinimumRevisionsToKeep = minimumRevisionsToKeep
}

func (c *RevisionsCollectionConfiguration) GetMinimumRevisionAgeToKeep() *time.Duration {
	return c.MinimumRevisionAgeToKeep
}

func (c *RevisionsCollectionConfiguration) SetMinimumRevisionAgeToKeep(minimumRevisionAgeToKeep *time.Duration) {
	c.MinimumRevisionAgeToKeep = minimumRevisionAgeToKeep
}

func (c *RevisionsCollectionConfiguration) IsDisabled() bool {
	return c.Disabled
}

func (c *RevisionsCollectionConfiguration) SetDisabled(disabled bool) {
	c.Disabled = disabled
}

func (c *RevisionsCollectionConfiguration) IsPurgeOnDelete() bool {
	return c.PurgeOnDelete
}

func (c *RevisionsCollectionConfiguration) SetPurgeOnDelete(purgeOnDelete bool) {
	c.PurgeOnDelete = purgeOnDelete
}
