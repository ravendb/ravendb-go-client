package ravendb

import "reflect"

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

// documentInfo stores information about entity in a session
// TODO: maybe route all places where we compare enity for equality via
// documentInfo.Equal(other interface{}), so that we can catch
// mismatches of *struct and **struct
type documentInfo struct {
	id                   string
	changeVector         *string
	concurrencyCheckMode ConcurrencyCheckMode
	ignoreChanges        bool
	metadata             ObjectNode
	document             ObjectNode
	metadataInstance     *MetadataAsDictionary
	entity               interface{}
	newDocument          bool
	collection           string
}

// we want to route assignments to entity through this functions
// so that we can maintain invariant that entity is *struct (and
// not, e.g., **struct). It's hard to track the difference between
// *struct and **struct otherwise
func (d *documentInfo) setEntity(value interface{}) {
	tp := reflect.TypeOf(value)
	if tp.Kind() == reflect.Struct {
		panicIf(true, "trying to set struct %T", value)
		d.entity = value
	}

	if tp.Kind() != reflect.Ptr || tp.Elem() == nil {
		//panicIf(tp.Kind() != reflect.Ptr || tp.Elem() == nil, "expected value to be *struct or **struct, is %T", value)
		//TODO: re-enable this panic and fix places that trigger it
		d.entity = value
		return
	}
	tp = tp.Elem()
	if tp.Kind() == reflect.Struct {
		// if it's *struct, just assign
		d.entity = value
		return
	}
	if tp.Kind() != reflect.Ptr || tp.Elem() == nil || tp.Elem().Kind() != reflect.Struct {
		//panicIf(tp.Kind() != reflect.Ptr || tp.Elem() == nil || tp.Elem().Kind() != reflect.Struct, "expected value to be *struct or **struct, is %T", value)
		//TODO: re-enable this panic and fix places that trigger it
		d.entity = value
		return

	}
	// it's **struct, so extract *struct
	rv := reflect.ValueOf(value)
	rv = rv.Elem() // it's *struct now
	d.entity = rv.Interface()
}

func getNewDocumentInfo(document ObjectNode) *documentInfo {
	metadataV, ok := document[MetadataKey]
	// TODO: maybe convert to errors
	panicIf(!ok, "Document must have a metadata")
	metadata, ok := metadataV.(ObjectNode)
	panicIf(!ok, "Document metadata is not a valid type %T", metadataV)

	// TODO: return an error?

	id, ok := JsonGetAsText(metadata, MetadataID)
	// TODO: return an error?
	panicIf(!ok || id == "", "Document must have an id")

	changeVector := jsonGetAsTextPointer(metadata, MetadataChangeVector)
	// TODO: return an error?
	panicIf(changeVector == nil, "Document must have a Change Vector")

	newDocumentInfo := &documentInfo{}
	newDocumentInfo.id = id
	newDocumentInfo.document = document
	newDocumentInfo.metadata = metadata

	newDocumentInfo.changeVector = changeVector
	return newDocumentInfo
}
