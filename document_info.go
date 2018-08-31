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
	changeVector         *string
	concurrencyCheckMode ConcurrencyCheckMode
	ignoreChanges        bool
	metadata             ObjectNode
	document             ObjectNode
	metadataInstance     *IMetadataDictionary
	entity               interface{}
	newDocument          bool
	collection           string
}

func NewDocumentInfo() *DocumentInfo {
	return &DocumentInfo{}
}

// TODO: remove those functions. Those are only to make porting faster, initially

func (d *DocumentInfo) setMetadataInstance(metadataInstance *IMetadataDictionary) {
	d.metadataInstance = metadataInstance
}

func (d *DocumentInfo) isNewDocument() bool {
	return d.newDocument
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

func (d *DocumentInfo) setChangeVector(changeVector *string) {
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
	// TODO: maybe convert to errors
	panicIf(!ok, "Document must have a metadata")
	metadata, ok := metadataV.(ObjectNode)
	panicIf(!ok, "Document metadata is not a valid type %T", metadataV)

	// TODO: return an error?

	id, ok := JsonGetAsText(metadata, Constants_Documents_Metadata_ID)
	// TODO: return an error?
	panicIf(!ok || id == "", "Document must have an id")

	changeVector := jsonGetAsTextPointer(metadata, Constants_Documents_Metadata_CHANGE_VECTOR)
	// TODO: return an error?
	panicIf(changeVector == nil, "Document must have a Change Vector")

	newDocumentInfo := NewDocumentInfo()
	newDocumentInfo.setId(id)
	newDocumentInfo.setDocument(document)
	newDocumentInfo.setMetadata(metadata)
	newDocumentInfo.setEntity(nil)
	newDocumentInfo.setChangeVector(changeVector)
	return newDocumentInfo
}
