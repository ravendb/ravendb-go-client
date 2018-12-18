package ravendb

import "io"

type PutAttachmentCommandData struct {
	*CommandData
	stream      io.Reader
	contentType string
}

var _ ICommandData = &PutAttachmentCommandData{} // verify interface match

func NewPutAttachmentCommandData(documentID string, name string, stream io.Reader, contentType string, changeVector *string) *PutAttachmentCommandData {
	panicIf(documentID == "", "DocumentId cannot be empty")
	panicIf(name == "", "Name cannot be empty")

	res := &PutAttachmentCommandData{
		CommandData: &CommandData{
			Type:         CommandType_ATTACHMENT_PUT,
			ID:           documentID,
			Name:         name,
			ChangeVector: changeVector,
		},
		stream:      stream,
		contentType: contentType,
	}
	return res
}

func (d *PutAttachmentCommandData) serialize(conventions *DocumentConventions) (interface{}, error) {
	res := d.baseJSON()
	res["Name"] = d.Name
	res["ContentType"] = d.contentType
	res["Type"] = "AttachmentPUT"
	res["ChangeVector"] = d.ChangeVector
	return res, nil
}

func (d *PutAttachmentCommandData) getStream() io.Reader {
	return d.stream
}

func (d *PutAttachmentCommandData) GetContentType() string {
	return d.contentType
}
