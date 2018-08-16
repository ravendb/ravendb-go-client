package ravendb

type PutCommandDataWithJson struct {
	*CommandData
	document ObjectNode
}

var _ ICommandData = &PutCommandDataWithJson{} // verify interface match

func NewPutCommandDataWithJson(id string, changeVector *string, document ObjectNode) *PutCommandDataWithJson {
	panicIf(document == nil, "Document cannot be nil")

	res := &PutCommandDataWithJson{
		CommandData: &CommandData{
			Type:         CommandType_PUT,
			ID:           id,
			ChangeVector: changeVector,
		},
		document: document,
	}
	return res
}

func (d *PutCommandDataWithJson) serialize(conventions *DocumentConventions) (interface{}, error) {
	js := d.baseJSON()
	js["Document"] = d.document
	return js, nil
}
