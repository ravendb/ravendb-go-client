package ravendb

// ConcurrencyCheckMode describes concurrency check
type ConcurrencyCheckMode int

const (
	// ConcurrencyCheckAuto is automatic optimistic concurrency check depending on UseOptimisticConcurrency setting or provided Change Vector
	ConcurrencyCheck_AUTO ConcurrencyCheckMode = iota
	// ConcurrencyCheckForced forces optimistic concurrency check even if UseOptimisticConcurrency is not set
	ConcurrencyCheck_FORCED
	// ConcurrencyCheckDisabled disables optimistic concurrency check even if UseOptimisticConcurrency is set
	ConcurrencyCheck_DISABLED
)

// DocumentInfo stores information about entity in a session
type DocumentInfo struct {
	id                   string
	changeVector         string
	concurrencyCheckMode ConcurrencyCheckMode
	ignoreChanges        bool
	metadata             ObjectNode
	document             ObjectNode
	metadataInstance     *IMetadataDictionary
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

func (d *DocumentInfo) isIgnoreChanges() bool {
	return d.ignoreChanges
}

func (d *DocumentInfo) getChangeVector() string {
	return d.changeVector
}

func (d *DocumentInfo) getMetadata() ObjectNode {
	return d.metadata
}

func (d *DocumentInfo) getMetadataInstance() *IMetadataDictionary {
	return d.metadataInstance
}

func (d *DocumentInfo) getConcurrencyCheckMode() ConcurrencyCheckMode {
	return d.concurrencyCheckMode
}

func (d *DocumentInfo) setMetadataInstance(metadataInstance *IMetadataDictionary) {
	d.metadataInstance = metadataInstance
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

func (d *DocumentInfo) setIgnoreChanges(ignoreChanges bool) {
	d.ignoreChanges = ignoreChanges
}

func DocumentInfo_getNewDocumentInfo(document ObjectNode) *DocumentInfo {
	metadataV, ok := document[Constants_Documents_Metadata_KEY]
	panicIf(!ok, "Document must have a metadata")
	metadata, ok := metadataV.(ObjectNode)
	panicIf(!ok, "Document metadata is not a valid type %T", metadataV)

	// TODO: return an error?

	id := jsonGetAsText(metadata, Constants_Documents_Metadata_ID)
	panicIf(id == "", "Document must have an id")
	// TODO: return an error?

	changeVector := jsonGetAsText(metadata, Constants_Documents_Metadata_CHANGE_VECTOR)
	panicIf(id == "", "Document must have a Change Vector")
	// TODO: return an error?

	newDocumentInfo := NewDocumentInfo()
	newDocumentInfo.setId(id)
	newDocumentInfo.setDocument(document)
	newDocumentInfo.setMetadata(metadata)
	newDocumentInfo.setEntity(nil)
	newDocumentInfo.setChangeVector(changeVector)
	return newDocumentInfo
}
