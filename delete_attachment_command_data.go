package ravendb

var _ ICommandData = &DeleteAttachmentCommandData{}

type DeleteAttachmentCommandData struct {
	CommandData
}

func NewDeleteAttachmentCommandData(documentId string, name string, destinationDocumentId string, changeVector string) (*DeleteAttachmentCommandData, error) {
	if stringIsWhitespace(documentId) {
		return nil, newIllegalArgumentError("DocumentId cannot be empty")
	}

	if stringIsWhitespace(name) {
		return nil, newIllegalArgumentError("Name cannot be empty")
	}

	res := &DeleteAttachmentCommandData{
		CommandData: CommandData{
			ID:           documentId,
			Name:         name,
			ChangeVector: stringToPtr(changeVector),
			Type:         CommandAttachmentDelete,
		},
	}
	return res, nil
}

func (d *DeleteAttachmentCommandData) serialize(conventions *DocumentConventions) (interface{}, error) {
	js := d.baseJSON()
	if d.Name != "" {
		js["Name"] = d.Name
	}
	return js, nil
}
