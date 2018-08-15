package ravendb

type RevisionsConfiguration struct {
	DefaultConfig *RevisionsCollectionConfiguration `json:"Default"`

	Collections map[string]*RevisionsCollectionConfiguration `json:"Collections"`
}

func NewRevisionsConfiguration() *RevisionsConfiguration {
	return &RevisionsConfiguration{}
}

func (c *RevisionsConfiguration) GetDefaultConfig() *RevisionsCollectionConfiguration {
	return c.DefaultConfig
}

func (c *RevisionsConfiguration) SetDefaultConfig(defaultConfig *RevisionsCollectionConfiguration) {
	c.DefaultConfig = defaultConfig
}

func (c *RevisionsConfiguration) GetCollections() map[string]*RevisionsCollectionConfiguration {
	return c.Collections
}

func (c *RevisionsConfiguration) SetCollections(collections map[string]*RevisionsCollectionConfiguration) {
	c.Collections = collections
}
