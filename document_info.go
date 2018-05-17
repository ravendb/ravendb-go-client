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
