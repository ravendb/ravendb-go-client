package ravendb

// DatabaseStatistics describes a result of GetStatisticsCommand
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/documents/operations/DatabaseStatistics.java#L8:14
type DatabaseStatistics struct {
	LastDocEtag               int64 `json:"LastDocEtag"`
	CountOfIndexes            int64 `json:"CountOfIndexes"`
	CountOfDocuments          int64 `json:"CountOfDocuments"`
	CountOfRevisionDocuments  int64 `json:"CountOfRevisionDocuments"` // TODO: present in Java, not seen in JSON
	CountOfDocumentsConflicts int64 `json:"CountOfDocumentsConflicts"`
	CountOfTombstones         int64 `json:"CountOfTombstones"`
	CountOfConflicts          int64 `json:"CountOfConflicts"`
	CountOfAttachments        int64 `json:"CountOfAttachments"`
	CountOfUniqueAttachments  int64 `json:"CountOfUniqueAttachments"`

	Indexes []IndexInformation `json:"Indexes"`

	DatabaseChangeVector                     string      `json:"DatabaseChangeVector"`
	DatabaseID                               string      `json:"DatabaseId"`
	Is64Bit                                  bool        `json:"Is64Bit"`
	Pager                                    string      `json:"Pager"`
	LastIndexingTime                         interface{} `json:"LastIndexingTime"` // TODO: this is time, can be null so must be a pointer
	SizeOnDisk                               *Size       `json:"SizeOnDisk"`
	NumberOfTransactionMergerQueueOperations int64       `json:"NumberOfTransactionMergerQueueOperations"`
}

/*
public IndexInformation[] getStaleIndexes() {
	return Arrays.stream(indexes)
		.filter(x -> x.isStale())
		.toArray(IndexInformation[]::new);
}

public IndexInformation[] getIndexes() {
	return indexes;
}

public void setIndexes(IndexInformation[] indexes) {
	this.indexes = indexes;
}

public Long getLastDocEtag() {
	return lastDocEtag;
}

public void setLastDocEtag(Long lastDocEtag) {
	this.lastDocEtag = lastDocEtag;
}

public int getCountOfIndexes() {
	return countOfIndexes;
}

public void setCountOfIndexes(int countOfIndexes) {
	this.countOfIndexes = countOfIndexes;
}

public long getCountOfDocuments() {
	return countOfDocuments;
}

public void setCountOfDocuments(long countOfDocuments) {
	this.countOfDocuments = countOfDocuments;
}

public long getCountOfRevisionDocuments() {
	return countOfRevisionDocuments;
}

public void setCountOfRevisionDocuments(long countOfRevisionDocuments) {
	this.countOfRevisionDocuments = countOfRevisionDocuments;
}

public long getCountOfDocumentsConflicts() {
	return countOfDocumentsConflicts;
}

public void setCountOfDocumentsConflicts(long countOfDocumentsConflicts) {
	this.countOfDocumentsConflicts = countOfDocumentsConflicts;
}

public long getCountOfTombstones() {
	return countOfTombstones;
}

public void setCountOfTombstones(long countOfTombstones) {
	this.countOfTombstones = countOfTombstones;
}

public long getCountOfConflicts() {
	return countOfConflicts;
}

public void setCountOfConflicts(long countOfConflicts) {
	this.countOfConflicts = countOfConflicts;
}

public long getCountOfAttachments() {
	return countOfAttachments;
}

public void setCountOfAttachments(long countOfAttachments) {
	this.countOfAttachments = countOfAttachments;
}

public long getCountOfUniqueAttachments() {
	return countOfUniqueAttachments;
}

public void setCountOfUniqueAttachments(long countOfUniqueAttachments) {
	this.countOfUniqueAttachments = countOfUniqueAttachments;
}

public string getDatabaseChangeVector() {
	return databaseChangeVector;
}

public void setDatabaseChangeVector(string databaseChangeVector) {
	this.databaseChangeVector = databaseChangeVector;
}

public string getDatabaseId() {
	return databaseId;
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

public string getPager() {
	return pager;
}

public void setPager(string pager) {
	this.pager = pager;
}

public Date getLastIndexingTime() {
	return lastIndexingTime;
}

public void setLastIndexingTime(Date lastIndexingTime) {
	this.lastIndexingTime = lastIndexingTime;
}


public Size getSizeOnDisk() {
	return sizeOnDisk;
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
