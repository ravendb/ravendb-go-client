package ravendb

type PatchStatus = string

const (
	PatchStatus_DOCUMENT_DOES_NOT_EXIST = "DocumentDoesNotExist"
	PatchStatus_CREATED                 = "Created"
	PatchStatus_PATCHED                 = "Patched"
	PatchStatus_SKIPPED                 = "Skipped"
	PatchStatus_NOT_MODIFIED            = "NotModified"
)
