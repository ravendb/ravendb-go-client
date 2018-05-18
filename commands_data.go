package ravendb

// CommandData describes data for a command
type CommandData struct {
	id           string
	typ          string
	name         string
	changeVector string
	document     map[string]interface{}
	json         map[string]interface{}
}

func (d *CommandData) getId() string {
	return d.id
}

func (d *CommandData) getType() string {
	return d.typ
}

func (d *CommandData) baseJSON() ObjectNode {
	res := ObjectNode{
		"Id":   d.id,
		"Type": d.typ,
	}
	// TODO: send null whnn empty?
	if d.changeVector != "" {
		res["ChangeVector"] = d.changeVector
	}
	return res
}

// NewPutCommandData creates CommandData for Put command
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/commands/commands_data.py#L22
//
func NewPutCommandData(id string, changeVector string, document map[string]interface{}, metadata map[string]interface{}) *CommandData {
	panicIf(document == nil, "document can't be nil")
	res := &CommandData{
		id:           id,
		typ:          CommandType_PUT,
		changeVector: changeVector,
		document:     document,
	}
	res.json = res.baseJSON()
	if metadata != nil {
		document["@metadata"] = metadata
	}
	res.json["Document"] = document
	return res
}

func NewPutCommandDataWithJson(id string, changeVector string, document ObjectNode) *CommandData {
	return NewPutCommandData(id, changeVector, document, nil)
}

// NewDeleteCommandData creates CommandData for Delete command
func NewDeleteCommandData(id string, changeVector string) *CommandData {
	res := &CommandData{
		id:           id,
		typ:          CommandType_DELETE,
		changeVector: changeVector,
	}
	res.json = res.baseJSON()
	return res
}

// TODO: PatchCommandData
// TODO: PutAttachmentCommandData
// TODO: DeleteAttachmentCommandData
