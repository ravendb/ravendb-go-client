package ravendb

type IndexStats struct {
	Name                          string  `json:"Name"`
	MapAttempts                   int     `json:"MapAttempts"`
	MapSuccesses                  int     `json:"MapSuccesses"`
	MapErrors                     int     `json:"MapErrors"`
	ReduceAttempts                *int    `json:"ReduceAttempts"`
	ReduceSuccesses               *int    `json:"ReduceSuccesses"`
	ReduceErrors                  *int    `json:"ReduceErrors"`
	MappedPerSecondRate           float64 `json:"MappedPerSecondRate"`
	ReducedPerSecondRate          float64 `json:"ReducedPerSecondRate"`
	MaxNumberOfOutputsPerDocument int     `json:"MaxNumberOfOutputsPerDocument"`

	Collections map[string]*CollectionStats `json:"Collections"`

	LastQueryingTime ServerTime         `json:"LastQueryingTime"`
	State            IndexState         `json:"State"`
	Priority         IndexPriority      `json:"Priority"`
	CreatedTimestamp ServerTime         `json:"CreatedTimestamp"`
	LastIndexingTime ServerTime         `json:"LastIndexingTime"`
	IsStale          bool               `json:"Stale"`
	LockMode         IndexLockMode      `json:"LockMode"`
	Type             IndexType          `json:"Type"`
	Status           IndexRunningStatus `json:"Status"`
	EntriesCount     int                `json:"EntriesCount"`
	ErrorsCount      int                `json:"ErrorsCount"`
	TestIndex        bool               `json:"IsTestIndex"`
}

type CollectionStats struct {
	LastProcessedDocumentEtag  int64 `json:"LastProcessedDocumentEtag"`
	LastProcessedTombstoneEtag int64 `json:"LastProcessedTombstoneEtag"`
	DocumentLag                int64 `json:"DocumentLag"`
	TombstoneLag               int64 `json:"TombstoneLag"`
}

func NewCollectionStats() *CollectionStats {
	return &CollectionStats{
		DocumentLag:  -1,
		TombstoneLag: -1,
	}
}
