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
	Stale            bool               `json:"Stale"`
	LockMode         IndexLockMode      `json:"LockMode"`
	Type             IndexType          `json:"Type"`
	Status           IndexRunningStatus `json:"Status"`
	EntriesCount     int                `json:"EntriesCount"`
	ErrorsCount      int                `json:"ErrorsCount"`
	IsTestIndex      bool               `json:"IsTestIndex"`
}

func (s *IndexStats) GetName() string {
	return s.Name
}

func (s *IndexStats) setName(name string) {
	s.Name = name
}

func (s *IndexStats) getMapAttempts() int {
	return s.MapAttempts
}

func (s *IndexStats) setMapAttempts(mapAttempts int) {
	s.MapAttempts = mapAttempts
}

func (s *IndexStats) getMapSuccesses() int {
	return s.MapSuccesses
}

func (s *IndexStats) setMapSuccesses(mapSuccesses int) {
	s.MapSuccesses = mapSuccesses
}

func (s *IndexStats) getMapErrors() int {
	return s.MapErrors
}

func (s *IndexStats) setMapErrors(mapErrors int) {
	s.MapErrors = mapErrors
}

func (s *IndexStats) getReduceAttempts() *int {
	return s.ReduceAttempts
}

func (s *IndexStats) setReduceAttempts(reduceAttempts *int) {
	s.ReduceAttempts = reduceAttempts
}

func (s *IndexStats) getReduceSuccesses() *int {
	return s.ReduceSuccesses
}

func (s *IndexStats) setReduceSuccesses(reduceSuccesses *int) {
	s.ReduceSuccesses = reduceSuccesses
}

func (s *IndexStats) getReduceErrors() *int {
	return s.ReduceErrors
}

func (s *IndexStats) setReduceErrors(reduceErrors *int) {
	s.ReduceErrors = reduceErrors
}

func (s *IndexStats) getMappedPerSecondRate() float64 {
	return s.MappedPerSecondRate
}

func (s *IndexStats) setMappedPerSecondRate(mappedPerSecondRate float64) {
	s.MappedPerSecondRate = mappedPerSecondRate
}

func (s *IndexStats) getReducedPerSecondRate() float64 {
	return s.ReducedPerSecondRate
}

func (s *IndexStats) setReducedPerSecondRate(reducedPerSecondRate float64) {
	s.ReducedPerSecondRate = reducedPerSecondRate
}

func (s *IndexStats) getMaxNumberOfOutputsPerDocument() int {
	return s.MaxNumberOfOutputsPerDocument
}

func (s *IndexStats) setMaxNumberOfOutputsPerDocument(maxNumberOfOutputsPerDocument int) {
	s.MaxNumberOfOutputsPerDocument = maxNumberOfOutputsPerDocument
}

func (s *IndexStats) getCollections() map[string]*CollectionStats {
	return s.Collections
}

func (s *IndexStats) setCollections(collections map[string]*CollectionStats) {
	s.Collections = collections
}

func (s *IndexStats) getLastQueryingTime() ServerTime {
	return s.LastQueryingTime
}

func (s *IndexStats) setLastQueryingTime(lastQueryingTime ServerTime) {
	s.LastQueryingTime = lastQueryingTime
}

func (s *IndexStats) getState() IndexState {
	return s.State
}

func (s *IndexStats) setState(state IndexState) {
	s.State = state
}

func (s *IndexStats) getPriority() IndexPriority {
	return s.Priority
}

func (s *IndexStats) setPriority(priority IndexPriority) {
	s.Priority = priority
}

func (s *IndexStats) getCreatedTimestamp() ServerTime {
	return s.CreatedTimestamp
}

func (s *IndexStats) setCreatedTimestamp(createdTimestamp ServerTime) {
	s.CreatedTimestamp = createdTimestamp
}

func (s *IndexStats) getLastIndexingTime() ServerTime {
	return s.LastIndexingTime
}

func (s *IndexStats) setLastIndexingTime(lastIndexingTime ServerTime) {
	s.LastIndexingTime = lastIndexingTime
}

func (s *IndexStats) isStale() bool {
	return s.Stale
}

func (s *IndexStats) setStale(stale bool) {
	s.Stale = stale
}

func (s *IndexStats) getLockMode() IndexLockMode {
	return s.LockMode
}

func (s *IndexStats) setLockMode(lockMode IndexLockMode) {
	s.LockMode = lockMode
}

func (s *IndexStats) getType() IndexType {
	return s.Type
}

func (s *IndexStats) setType(typ IndexType) {
	s.Type = typ
}

func (s *IndexStats) getStatus() IndexRunningStatus {
	return s.Status
}

func (s *IndexStats) setStatus(status IndexRunningStatus) {
	s.Status = status
}

func (s *IndexStats) getEntriesCount() int {
	return s.EntriesCount
}

func (s *IndexStats) setEntriesCount(entriesCount int) {
	s.EntriesCount = entriesCount
}

func (s *IndexStats) getErrorsCount() int {
	return s.ErrorsCount
}

func (s *IndexStats) setErrorsCount(errorsCount int) {
	s.ErrorsCount = errorsCount
}

func (s *IndexStats) isTestIndex() bool {
	return s.IsTestIndex
}

func (s *IndexStats) setTestIndex(testIndex bool) {
	s.IsTestIndex = testIndex
}

type CollectionStats struct {
	LastProcessedDocumentEtag  int `json:"LastProcessedDocumentEtag"`
	LastProcessedTombstoneEtag int `json:"LastProcessedTombstoneEtag"`
	DocumentLag                int `json:"DocumentLag"`
	TombstoneLag               int `json:"TombstoneLag"`
}

func NewCollectionStats() *CollectionStats {
	return &CollectionStats{
		DocumentLag:  -1,
		TombstoneLag: -1,
	}
}

func (s *CollectionStats) getLastProcessedDocumentEtag() int {
	return s.LastProcessedDocumentEtag
}

func (s *CollectionStats) setLastProcessedDocumentEtag(lastProcessedDocumentEtag int) {
	s.LastProcessedDocumentEtag = lastProcessedDocumentEtag
}

func (s *CollectionStats) getLastProcessedTombstoneEtag() int {
	return s.LastProcessedTombstoneEtag
}

func (s *CollectionStats) setLastProcessedTombstoneEtag(lastProcessedTombstoneEtag int) {
	s.LastProcessedTombstoneEtag = lastProcessedTombstoneEtag
}

func (s *CollectionStats) getDocumentLag() int {
	return s.DocumentLag
}

func (s *CollectionStats) setDocumentLag(documentLag int) {
	s.DocumentLag = documentLag
}

func (s *CollectionStats) getTombstoneLag() int {
	return s.TombstoneLag
}

func (s *CollectionStats) setTombstoneLag(tombstoneLag int) {
	s.TombstoneLag = tombstoneLag
}
