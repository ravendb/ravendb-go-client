package ravendb

var _ ICommandData = &CopyAttachmentCommandData{}

type CopyAttachmentCommandData struct {
	CommandData

	destinationId   string
	destinationName string
}

func NewCopyAttachmentCommandData(sourceDocumentId string, sourceName string, destinationDocumentId string, destinationName string, changeVector string) (*CopyAttachmentCommandData, error) {
	if stringIsWhitespace(sourceDocumentId) {
		return nil, newIllegalArgumentError("SourceDocumentId is required")
	}

	if stringIsWhitespace(sourceName) {
		return nil, newIllegalArgumentError("SourceName is required")
	}

	if stringIsWhitespace(destinationDocumentId) {
		return nil, newIllegalArgumentError("DestinationDocumentId is required")
	}

	if stringIsWhitespace(destinationName) {
		return nil, newIllegalArgumentError("DestinationName is required")
	}

	res := &CopyAttachmentCommandData{
		CommandData: CommandData{
			ID:           sourceDocumentId,
			Name:         sourceName,
			ChangeVector: stringToPtr(changeVector),
			Type:         CommandAttachmentCopy,
		},
		destinationId:   destinationDocumentId,
		destinationName: destinationName,
	}
	return res, nil
}

func (d *CopyAttachmentCommandData) serialize(conventions *DocumentConventions) (interface{}, error) {
	js := d.baseJSON()
	if d.Name != "" {
		js["Name"] = d.Name
	}
	js["DestinationId"] = d.destinationId
	js["DestinationName"] = d.destinationName
	return js, nil
}
