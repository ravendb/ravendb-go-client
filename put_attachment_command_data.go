package ravendb

import "io"

type PutAttachmentCommandData struct {
	*CommandData
	stream      io.Reader
	contentType string
}

var _ ICommandData = &PutAttachmentCommandData{} // verify interface match

func NewPutAttachmentCommandData(documentID string, name string, stream io.Reader, contentType string, changeVector *string) (*PutAttachmentCommandData, error) {
	if stringIsBlank(documentID) {
		return nil, newIllegalArgumentError("DocumentId cannot be null or empty")
	}
	if stringIsBlank(name) {
		return nil, newIllegalArgumentError("Name cannot be null or empty")
	}

	res := &PutAttachmentCommandData{
		CommandData: &CommandData{
			Type:         CommandAttachmentPut,
			ID:           documentID,
			Name:         name,
			ChangeVector: changeVector,
		},
		stream:      stream,
		contentType: contentType,
	}
	return res, nil
}

func (d *PutAttachmentCommandData) serialize(conventions *DocumentConventions) (interface{}, error) {
	res := d.baseJSON()
	res["Name"] = d.Name
	if d.contentType != "" {
		res["ContentType"] = d.contentType
	} else {
		res["ContentType"] = nil
	}
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
