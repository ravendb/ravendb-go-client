package ravendb

type DeleteAttachmentCommandData struct {
	*CommandData
}

// NewDeleteAttachmentCommandData creates CommandData for Delete Attachment command
// TODO: return a concrete type?
func NewDeleteAttachmentCommandData(documentId string, name string, changeVector *string) ICommandData {
	res := &DeleteAttachmentCommandData{
		&CommandData{
			Type:         CommandType_DELETE,
			ID:           documentId,
			Name:         name,
			ChangeVector: changeVector,
		},
	}
	return res
}

func (d *DeleteAttachmentCommandData) serialize(conventions *DocumentConventions) (interface{}, error) {
	res := d.baseJSON()
	res["Type"] = "AttachmentDELETE"
	return res, nil
}
