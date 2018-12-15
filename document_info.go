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

// documentInfo stores information about entity in a session
type documentInfo struct {
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

func getNewDocumentInfo(document ObjectNode) *documentInfo {
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

	newDocumentInfo := &documentInfo{}
	newDocumentInfo.id = id
	newDocumentInfo.document = document
	newDocumentInfo.metadata = metadata

	newDocumentInfo.changeVector = changeVector
	return newDocumentInfo
}
