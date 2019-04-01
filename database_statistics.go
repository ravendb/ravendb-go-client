package ravendb

import "time"

// DatabaseStatistics describes a result of GetStatisticsCommand
type DatabaseStatistics struct {
	LastDocEtag               int64 `json:"LastDocEtag"`
	CountOfIndexes            int   `json:"CountOfIndexes"`
	CountOfDocuments          int64 `json:"CountOfDocuments"`
	CountOfRevisionDocuments  int64 `json:"CountOfRevisionDocuments"`
	CountOfDocumentsConflicts int64 `json:"CountOfDocumentsConflicts"`
	CountOfTombstones         int64 `json:"CountOfTombstones"`
	CountOfConflicts          int64 `json:"CountOfConflicts"`
	CountOfAttachments        int64 `json:"CountOfAttachments"`
	CountOfCounters           int64 `json:"CountOfCounters"`
	CountOfUniqueAttachments  int64 `json:"CountOfUniqueAttachments"`

	Indexes []*IndexInformation `json:"Indexes"`

	DatabaseChangeVector                     string `json:"DatabaseChangeVector"`
	DatabaseID                               string `json:"DatabaseId"`
	Is64Bit                                  bool   `json:"Is64Bit"`
	Pager                                    string `json:"Pager"`
	LastIndexingTime                         *Time  `json:"LastIndexingTime"`
	SizeOnDisk                               *Size  `json:"SizeOnDisk"`
	TempBuffersSizeOnDisk                    *Size  `json:"TempBuffersSizeOnDisk"`
	NumberOfTransactionMergerQueueOperations int    `json:"NumberOfTransactionMergerQueueOperations"`
}

// GetLastIndexingTime returns last indexing time
func (s *DatabaseStatistics) GetLastIndexingTime() *time.Time {
	return s.LastIndexingTime.toTimePtr()
}

/*
public IndexInformation[] getStaleIndexes() {
	return Arrays.stream(indexes)
		.filter(x -> x.isStale())
		.toArray(IndexInformation[]::new);
}
*/
