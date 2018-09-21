package ravendb

import "time"

// DatabaseStatistics describes a result of GetStatisticsCommand
type DatabaseStatistics struct {
	LastDocEtag               int `json:"LastDocEtag"`
	CountOfIndexes            int `json:"CountOfIndexes"`
	CountOfDocuments          int `json:"CountOfDocuments"`
	CountOfRevisionDocuments  int `json:"CountOfRevisionDocuments"` // TODO: present in Java, not seen in JSON
	CountOfDocumentsConflicts int `json:"CountOfDocumentsConflicts"`
	CountOfTombstones         int `json:"CountOfTombstones"`
	CountOfConflicts          int `json:"CountOfConflicts"`
	CountOfAttachments        int `json:"CountOfAttachments"`
	CountOfUniqueAttachments  int `json:"CountOfUniqueAttachments"`

	Indexes []*IndexInformation `json:"Indexes"`

	DatabaseChangeVector                     string      `json:"DatabaseChangeVector"`
	DatabaseID                               string      `json:"DatabaseId"`
	Is64Bit                                  bool        `json:"Is64Bit"`
	Pager                                    string      `json:"Pager"`
	LastIndexingTime                         *ServerTime `json:"LastIndexingTime"`
	SizeOnDisk                               *Size       `json:"SizeOnDisk"`
	NumberOfTransactionMergerQueueOperations int         `json:"NumberOfTransactionMergerQueueOperations"`
}

func (s *DatabaseStatistics) GetLastIndexingTime() *time.Time {
	return serverTimePtrToTimePtr(s.LastIndexingTime)
}

/*
public IndexInformation[] getStaleIndexes() {
	return Arrays.stream(indexes)
		.filter(x -> x.isStale())
		.toArray(IndexInformation[]::new);
}
*/
