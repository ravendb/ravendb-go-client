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
	metadata             ObjectNode
	document             ObjectNode
	metadataInstance     IMetadataDictionary
	entity               interface{}
	newDocument          bool
	collection           string

	// TODO: remove those, from python code, not Java
	originalMetadata map[string]interface{}
	originalValue    map[string]interface{}
}

func NewDocumentInfo() *DocumentInfo {
	return &DocumentInfo{}
}

// TODO: remove those functions. Those are only to make porting faster, initially
func (d *DocumentInfo) getId() string {
	return d.id
}

func (d *DocumentInfo) getEntity() interface{} {
	return d.entity
}

func (d *DocumentInfo) getChangeVector() string {
	return d.changeVector
}

func (d *DocumentInfo) setId(id string) {
	d.id = id
}

func (d *DocumentInfo) setNewDocument(isNew bool) {
	d.newDocument = isNew
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

func (d *DocumentInfo) setConcurrencyCheckMode(m ConcurrencyCheckMode) {
	d.concurrencyCheckMode = m
}
