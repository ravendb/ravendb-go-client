package ravendb

type RevisionsConfiguration struct {
	DefaultConfig *RevisionsCollectionConfiguration `json:"Default"`

	Collections map[string]*RevisionsCollectionConfiguration `json:"Collections"`
}

func NewRevisionsConfiguration() *RevisionsConfiguration {
	return &RevisionsConfiguration{}
}

func (c *RevisionsConfiguration) getDefaultConfig() *RevisionsCollectionConfiguration {
	return c.DefaultConfig
}

func (c *RevisionsConfiguration) setDefaultConfig(defaultConfig *RevisionsCollectionConfiguration) {
	c.DefaultConfig = defaultConfig
}

func (c *RevisionsConfiguration) getCollections() map[string]*RevisionsCollectionConfiguration {
	return c.Collections
}

func (c *RevisionsConfiguration) setCollections(collections map[string]*RevisionsCollectionConfiguration) {
	c.Collections = collections
}
