package ravendb

// CommandData describes data for a command
type CommandData struct {
	key          string
	typ          string
	name         string
	changeVector string
	document     map[string]interface{}
	json         map[string]interface{}
}

// NewPutCommandData creates CommandData for Put command
// https://sourcegraph.com/github.com/ravendb/RavenDB-Python-Client@v4.0/-/blob/pyravendb/commands/commands_data.py#L22
//
func NewPutCommandData(key string, changeVector string, document map[string]interface{}, metadata map[string]interface{}) *CommandData {
	panicIf(document == nil, "document can't be nil")
	res := &CommandData{
		key:          key,
		typ:          "PUT",
		changeVector: changeVector,
		document:     document,
	}
	if metadata != nil {
		document["@metadata"] = metadata
	}
	res.json = map[string]interface{}{
		"Type":     res.typ,
		"Id":       res.key,
		"Document": document,
	}
	if changeVector != "" {
		res.json["ChangeVector"] = changeVector
	}
	return res
}

// NewDeleteCommandData creates CommandData for Delete command
func NewDeleteCommandData(key string, changeVector string) *CommandData {
	res := &CommandData{
		key:          key,
		typ:          "DELETE",
		changeVector: changeVector,
	}
	res.json = map[string]interface{}{
		"Type": res.typ,
		"Id":   res.key,
	}
	if changeVector != "" {
		res.json["ChangeVector"] = changeVector
	}
	return res
}

// TODO: PatchCommandData
// TODO: PutAttachmentCommandData
// TODO: DeleteAttachmentCommandData
