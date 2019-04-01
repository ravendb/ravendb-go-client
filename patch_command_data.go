package ravendb

var _ ICommandData = &PatchCommandData{}

type PatchCommandData struct {
	*CommandData

	patch          *PatchRequest
	patchIfMissing *PatchRequest
	returnDocument bool
}

// NewPatchCommandData creates CommandData for Patch Attachment command
func NewPatchCommandData(id string, changeVector *string, patch *PatchRequest, patchIfMissing *PatchRequest) (*PatchCommandData, error) {
	if id == "" {
		return nil, newIllegalArgumentError("id cannot be empty")
	}
	if patch == nil {
		return nil, newIllegalArgumentError("Patch cannot be nil")
	}
	res := &PatchCommandData{
		CommandData: &CommandData{
			ID:           id,
			ChangeVector: changeVector,
			Type:         CommandPatch,
		},
		patch:          patch,
		patchIfMissing: patchIfMissing,
	}
	return res, nil
}

func (d *PatchCommandData) serialize(conventions *DocumentConventions) (interface{}, error) {
	res := d.baseJSON()
	res["Patch"] = d.patch.Serialize()

	if d.patchIfMissing != nil {
		res["PatchIfMissing"] = d.patchIfMissing.Serialize()
	}
	if d.returnDocument {
		res["ReturnDocument"] = d.returnDocument
	}
	return res, nil
}

func (d *PatchCommandData) onBeforeSaveChanges(session *InMemoryDocumentSessionOperations) {
	d.returnDocument = session.IsLoaded(d.ID)
}
