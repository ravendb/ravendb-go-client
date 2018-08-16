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
	TestIndex        bool               `json:"IsTestIndex"`
}

func (s *IndexStats) GetName() string {
	return s.Name
}

func (s *IndexStats) SetName(name string) {
	s.Name = name
}

func (s *IndexStats) GetMapAttempts() int {
	return s.MapAttempts
}

func (s *IndexStats) SetMapAttempts(mapAttempts int) {
	s.MapAttempts = mapAttempts
}

func (s *IndexStats) GetMapSuccesses() int {
	return s.MapSuccesses
}

func (s *IndexStats) SetMapSuccesses(mapSuccesses int) {
	s.MapSuccesses = mapSuccesses
}

func (s *IndexStats) GetMapErrors() int {
	return s.MapErrors
}

func (s *IndexStats) SetMapErrors(mapErrors int) {
	s.MapErrors = mapErrors
}

func (s *IndexStats) GetReduceAttempts() *int {
	return s.ReduceAttempts
}

func (s *IndexStats) SetReduceAttempts(reduceAttempts *int) {
	s.ReduceAttempts = reduceAttempts
}

func (s *IndexStats) GetReduceSuccesses() *int {
	return s.ReduceSuccesses
}

func (s *IndexStats) SetReduceSuccesses(reduceSuccesses *int) {
	s.ReduceSuccesses = reduceSuccesses
}

func (s *IndexStats) GetReduceErrors() *int {
	return s.ReduceErrors
}

func (s *IndexStats) SetReduceErrors(reduceErrors *int) {
	s.ReduceErrors = reduceErrors
}

func (s *IndexStats) GetMappedPerSecondRate() float64 {
	return s.MappedPerSecondRate
}

func (s *IndexStats) SetMappedPerSecondRate(mappedPerSecondRate float64) {
	s.MappedPerSecondRate = mappedPerSecondRate
}

func (s *IndexStats) GetReducedPerSecondRate() float64 {
	return s.ReducedPerSecondRate
}

func (s *IndexStats) SetReducedPerSecondRate(reducedPerSecondRate float64) {
	s.ReducedPerSecondRate = reducedPerSecondRate
}

func (s *IndexStats) GetMaxNumberOfOutputsPerDocument() int {
	return s.MaxNumberOfOutputsPerDocument
}

func (s *IndexStats) SetMaxNumberOfOutputsPerDocument(maxNumberOfOutputsPerDocument int) {
	s.MaxNumberOfOutputsPerDocument = maxNumberOfOutputsPerDocument
}

func (s *IndexStats) GetCollections() map[string]*CollectionStats {
	return s.Collections
}

func (s *IndexStats) SetCollections(collections map[string]*CollectionStats) {
	s.Collections = collections
}

func (s *IndexStats) GetLastQueryingTime() ServerTime {
	return s.LastQueryingTime
}

func (s *IndexStats) SetLastQueryingTime(lastQueryingTime ServerTime) {
	s.LastQueryingTime = lastQueryingTime
}

func (s *IndexStats) GetState() IndexState {
	return s.State
}

func (s *IndexStats) SetState(state IndexState) {
	s.State = state
}

func (s *IndexStats) GetPriority() IndexPriority {
	return s.Priority
}

func (s *IndexStats) SetPriority(priority IndexPriority) {
	s.Priority = priority
}

func (s *IndexStats) GetCreatedTimestamp() ServerTime {
	return s.CreatedTimestamp
}

func (s *IndexStats) SetCreatedTimestamp(createdTimestamp ServerTime) {
	s.CreatedTimestamp = createdTimestamp
}

func (s *IndexStats) GetLastIndexingTime() ServerTime {
	return s.LastIndexingTime
}

func (s *IndexStats) SetLastIndexingTime(lastIndexingTime ServerTime) {
	s.LastIndexingTime = lastIndexingTime
}

func (s *IndexStats) SsStale() bool {
	return s.Stale
}

func (s *IndexStats) SetStale(stale bool) {
	s.Stale = stale
}

func (s *IndexStats) GetLockMode() IndexLockMode {
	return s.LockMode
}

func (s *IndexStats) SetLockMode(lockMode IndexLockMode) {
	s.LockMode = lockMode
}

func (s *IndexStats) GetType() IndexType {
	return s.Type
}

func (s *IndexStats) SetType(typ IndexType) {
	s.Type = typ
}

func (s *IndexStats) GetStatus() IndexRunningStatus {
	return s.Status
}

func (s *IndexStats) SetStatus(status IndexRunningStatus) {
	s.Status = status
}

func (s *IndexStats) GetEntriesCount() int {
	return s.EntriesCount
}

func (s *IndexStats) SetEntriesCount(entriesCount int) {
	s.EntriesCount = entriesCount
}

func (s *IndexStats) GetErrorsCount() int {
	return s.ErrorsCount
}

func (s *IndexStats) SetErrorsCount(errorsCount int) {
	s.ErrorsCount = errorsCount
}

func (s *IndexStats) IsTestIndex() bool {
	return s.TestIndex
}

func (s *IndexStats) SetTestIndex(testIndex bool) {
	s.TestIndex = testIndex
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

func (s *CollectionStats) GetLastProcessedDocumentEtag() int {
	return s.LastProcessedDocumentEtag
}

func (s *CollectionStats) SetLastProcessedDocumentEtag(lastProcessedDocumentEtag int) {
	s.LastProcessedDocumentEtag = lastProcessedDocumentEtag
}

func (s *CollectionStats) GetLastProcessedTombstoneEtag() int {
	return s.LastProcessedTombstoneEtag
}

func (s *CollectionStats) SetLastProcessedTombstoneEtag(lastProcessedTombstoneEtag int) {
	s.LastProcessedTombstoneEtag = lastProcessedTombstoneEtag
}

func (s *CollectionStats) GetDocumentLag() int {
	return s.DocumentLag
}

func (s *CollectionStats) SetDocumentLag(documentLag int) {
	s.DocumentLag = documentLag
}

func (s *CollectionStats) GetTombstoneLag() int {
	return s.TombstoneLag
}

func (s *CollectionStats) SetTombstoneLag(tombstoneLag int) {
	s.TombstoneLag = tombstoneLag
}
