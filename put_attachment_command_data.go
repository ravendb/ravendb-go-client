package ravendb

import "io"

type PutAttachmentCommandData struct {
	*CommandData
	stream      io.Reader
	contentType string
}

var _ ICommandData = &PutAttachmentCommandData{} // verify interface match

func NewPutAttachmentCommandData(documentId string, name string, stream io.Reader, contentType string, changeVector *string) *PutAttachmentCommandData {
	panicIf(documentId == "", "DocumentId cannot be empty")
	panicIf(name == "", "Name cannot be empty")

	res := &PutAttachmentCommandData{
		CommandData: &CommandData{
			Type:         CommandType_ATTACHMENT_PUT,
			ID:           documentId,
			Name:         name,
			ChangeVector: changeVector,
		},
		stream:      stream,
		contentType: contentType,
	}
	return res
}

func (d *PutAttachmentCommandData) serialize(conventions *DocumentConventions) (interface{}, error) {
	// TODO: implement me
	panicIf(true, "NYI")
	return nil, nil
}

func (d *PutAttachmentCommandData) getStream() io.Reader {
	return d.stream
}

/*
public string getContentType() {
	return contentType;
}

@Override
public void serialize(JsonGenerator generator, DocumentConventions conventions) throws IOException {
	generator.writeStartObject();
	generator.writeStringField("Id", id);
	generator.writeStringField("Name", name);
	generator.writeStringField("ContentType", contentType);
	generator.writeStringField("ChangeVector", changeVector);
	generator.writeStringField("Type", "AttachmentPUT");
	generator.writeEndObject();
}
*/
