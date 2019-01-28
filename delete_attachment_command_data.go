package ravendb

type DeleteAttachmentCommandData struct {
	*CommandData
}

// NewDeleteAttachmentCommandData creates CommandData for Delete Attachment command
func NewDeleteAttachmentCommandData(documentID string, name string, changeVector *string) (*DeleteAttachmentCommandData, error) {
	if stringIsBlank(documentID) {
		return nil, newIllegalArgumentError("DocumentId cannot be null or empty")
	}
	if stringIsBlank(name) {
		return nil, newIllegalArgumentError("Name cannot be null or empty")
	}

	res := &DeleteAttachmentCommandData{
		&CommandData{
			Type:         CommandDelete,
			ID:           documentID,
			Name:         name,
			ChangeVector: changeVector,
		},
	}
	return res, nil
}

func (d *DeleteAttachmentCommandData) serialize(conventions *DocumentConventions) (interface{}, error) {
	res := d.baseJSON()
	res["Type"] = "AttachmentDELETE"
	res["Name"] = d.Name
	return res, nil
}
