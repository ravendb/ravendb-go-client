package ravendb

// IMetadataDictionary describes metadata for a document
type IMetadataDictionary = map[string]interface{}

// ConcurrencyCheckMode describes concurrency check
type ConcurrencyCheckMode int

const (
	// ConcurrencyCheckAuto is automatic optimistic concurrency check depending on UseOptimisticConcurrency setting or provided Change Vector
	ConcurrencyCheckAuto ConcurrencyCheckMode = iota
	// ConcurrencyCheckForced forces optimistic concurrency check even if UseOptimisticConcurrency is not set
	ConcurrencyCheckForced
	// ConcurrencyCheckDisabled disables optimistic concurrency check even if UseOptimisticConcurrency is set
	ConcurrencyCheckDisabled
)

// DocumentInfo stores information about entity in a session
type DocumentInfo struct {
	id                   string
	changeVector         string
	concurrencyCheckMode ConcurrencyCheckMode
	ignoreChanges        bool
	originalMetadata     map[string]interface{}
	metadata             ObjectNode
	document             ObjectNode
	originalValue        map[string]interface{}
	metadataInstance     IMetadataDictionary
	entity               interface{}
	newDocuemnt          bool
	collection           string
}

func NewDocumentInfo() *DocumentInfo {
	return &DocumentInfo{}
}

// TODO: remove those functions. Those are only to make porting faster, initially
func (d *DocumentInfo) getId() string {
	return d.id
}

func (d *DocumentInfo) setDocument(document ObjectNode) {
	d.document = document
}

func (d *DocumentInfo) setMetadata(metadata ObjectNode) {
	d.metadata = metadata
}

func (d *DocumentInfo) setEntity(entity interface{}) {
	d.entity = entity
}

func (d *DocumentInfo) setChangeVector(changeVector string) {
	d.changeVector = changeVector
}
