package ravendb

import "io"

var _ ICommandData = &PutAttachmentCommandData{}

type PutAttachmentCommandData struct {
	CommandData

	stream      io.Reader
	contentType string
}

func NewPutAttachmentCommandData(documentId string, name string, stream io.Reader, contentType string, changeVector string) (*PutAttachmentCommandData, error) {
	if stringIsWhitespace(documentId) {
		return nil, newIllegalArgumentError("DocumentId cannot be empty")
	}

	if stringIsWhitespace(name) {
		return nil, newIllegalArgumentError("Name cannot be empty")
	}

	res := &PutAttachmentCommandData{
		CommandData: CommandData{
			ID:           documentId,
			Name:         name,
			ChangeVector: stringToPtr(changeVector),
			Type:         CommandAttachmentPut,
		},
		stream:      stream,
		contentType: contentType,
	}
	return res, nil
}

func (d *PutAttachmentCommandData) serialize(conventions *DocumentConventions) (interface{}, error) {
	js := d.baseJSON()
	if d.Name != "" {
		js["Name"] = d.Name
	}
	js["ContentType"] = d.contentType
	return js, nil
}
