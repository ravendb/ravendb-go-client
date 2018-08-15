package ravendb

import "time"

// DatabaseStatistics describes a result of GetStatisticsCommand
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/documents/operations/DatabaseStatistics.java#L8:14
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

func (s *DatabaseStatistics) GetLastDocEtag() int {
	return s.LastDocEtag
}

func (s *DatabaseStatistics) GetCountOfIndexes() int {
	return s.CountOfIndexes
}

func (s *DatabaseStatistics) GetCountOfDocuments() int {
	return s.CountOfDocuments
}

func (s *DatabaseStatistics) GetCountOfRevisionDocuments() int {
	return s.CountOfRevisionDocuments
}

func (s *DatabaseStatistics) GetCountOfDocumentsConflicts() int {
	return s.CountOfDocumentsConflicts
}

func (s *DatabaseStatistics) GetCountOfTombstones() int {
	return s.CountOfTombstones
}

func (s *DatabaseStatistics) GetCountOfConflicts() int {
	return s.CountOfConflicts
}

func (s *DatabaseStatistics) GetCountOfAttachments() int {
	return s.CountOfAttachments
}

func (s *DatabaseStatistics) GetCountOfUniqueAttachments() int {
	return s.CountOfUniqueAttachments
}

func (s *DatabaseStatistics) GetDatabaseChangeVector() string {
	return s.DatabaseChangeVector
}

func (s *DatabaseStatistics) GetDatabaseID() string {
	return s.DatabaseID
}

func (s *DatabaseStatistics) GetPager() string {
	return s.Pager
}

func (s *DatabaseStatistics) GetLastIndexingTime() *time.Time {
	return serverTimePtrToTimePtr(s.LastIndexingTime)
}

func (s *DatabaseStatistics) GetIndexes() []*IndexInformation {
	return s.Indexes
}

func (s *DatabaseStatistics) GetSizeOnDisk() *Size {
	return s.SizeOnDisk
}

/*
public IndexInformation[] getStaleIndexes() {
	return Arrays.stream(indexes)
		.filter(x -> x.isStale())
		.toArray(IndexInformation[]::new);
}

public void setIndexes(IndexInformation[] indexes) {
	this.indexes = indexes;
}

public void setLastDocEtag(Long lastDocEtag) {
	this.lastDocEtag = lastDocEtag;
}

public void setCountOfIndexes(int countOfIndexes) {
	this.countOfIndexes = countOfIndexes;
}

public void setCountOfDocuments(long countOfDocuments) {
	this.countOfDocuments = countOfDocuments;
}

public void setCountOfRevisionDocuments(long countOfRevisionDocuments) {
	this.countOfRevisionDocuments = countOfRevisionDocuments;
}

public void setCountOfDocumentsConflicts(long countOfDocumentsConflicts) {
	this.countOfDocumentsConflicts = countOfDocumentsConflicts;
}

public void setCountOfTombstones(long countOfTombstones) {
	this.countOfTombstones = countOfTombstones;
}

public void setCountOfConflicts(long countOfConflicts) {
	this.countOfConflicts = countOfConflicts;
}


public void setCountOfAttachments(long countOfAttachments) {
	this.countOfAttachments = countOfAttachments;
}

public void setCountOfUniqueAttachments(long countOfUniqueAttachments) {
	this.countOfUniqueAttachments = countOfUniqueAttachments;
}

public void setDatabaseChangeVector(string databaseChangeVector) {
	this.databaseChangeVector = databaseChangeVector;
}

public void setDatabaseId(string databaseId) {
	this.databaseId = databaseId;
}

public boolean isIs64Bit() {
	return is64Bit;
}

public void setIs64Bit(boolean is64Bit) {
	this.is64Bit = is64Bit;
}

public void setPager(string pager) {
	this.pager = pager;
}

public void setLastIndexingTime(Date lastIndexingTime) {
	this.lastIndexingTime = lastIndexingTime;
}



public void setSizeOnDisk(Size sizeOnDisk) {
	this.sizeOnDisk = sizeOnDisk;
}

public int getNumberOfTransactionMergerQueueOperations() {
	return numberOfTransactionMergerQueueOperations;
}

public void setNumberOfTransactionMergerQueueOperations(int numberOfTransactionMergerQueueOperations) {
	this.numberOfTransactionMergerQueueOperations = numberOfTransactionMergerQueueOperations;
}
*/
