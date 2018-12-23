package ravendb

// PutCommandDataWithJSON represents data for put command with json
type PutCommandDataWithJSON struct {
	*CommandData
	document ObjectNode
}

var _ ICommandData = &PutCommandDataWithJSON{} // verify interface match

//NewPutCommandDataWithJSON returns new PutCommandDataWithJSON
func NewPutCommandDataWithJSON(id string, changeVector *string, document ObjectNode) *PutCommandDataWithJSON {
	panicIf(document == nil, "Document cannot be nil")

	res := &PutCommandDataWithJSON{
		CommandData: &CommandData{
			Type:         CommandPut,
			ID:           id,
			ChangeVector: changeVector,
		},
		document: document,
	}
	return res
}

func (d *PutCommandDataWithJSON) serialize(conventions *DocumentConventions) (interface{}, error) {
	js := d.baseJSON()
	js["Document"] = d.document
	return js, nil
}
