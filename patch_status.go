package ravendb

type PatchStatus = string

const (
	PatchStatusDocumentDoesNotExist = "DocumentDoesNotExist"
	PatchStatusCreated              = "Created"
	PatchStatusPatched              = "Patched"
	PatchStatusSkipped              = "Skipped"
	PatchStatusNotModified          = "NotModified"
)
