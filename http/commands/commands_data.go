package commands

import (
	"github.com/ravendb/ravendb-go-client/tools/types"
	"github.com/ravendb/ravendb-go-client/data"
)

type ICommandData interface{
	ToJson() map[string]interface{}
	GetCommand() bool
}

type commandData struct{
	Key string
	Metadata *data.Metadata
	ChangeVector []string
	AdditionalData interface{}
	command bool
}
func NewCommandData(key string, metadata data.Metadata, change_vector []string, additional_data interface{}) (*commandData, error){
	ref := &commandData{}
	ref.Key = key
	ref.Metadata = &metadata
	ref.ChangeVector = change_vector
	ref.AdditionalData = additional_data
	ref.command = true
	return ref, nil
}
func (ref *commandData) GetCommand() bool{
	return ref.command
}

type PutCommandData struct{
	commandData *commandData
	commandType string
	Document *types.Document
}
func NewPutCommandData(key string, metadata data.Metadata, change_vector []string, document *types.Document) (*PutCommandData, error){
	ref := &PutCommandData{}
	ref.commandData, _ = NewCommandData(key, metadata, change_vector, nil)
	ref.commandType = "PUT"
	ref.Document = document
	return ref, nil
}
func (ref *PutCommandData) ToJson() map[string]interface{}{
	ref.Document.Metadata = ref.commandData.Metadata
	out := make(map[string]interface{})
	out["Type"] = ref.commandType
	out["Id"] = ref.commandData.Key
	out["Document"] = *ref.Document
	out["ChangeVector"] = ref.commandData.ChangeVector
	return out
}
func (ref *PutCommandData) GetCommand() bool{
	return ref.commandData.GetCommand()
}
type DeleteCommandData struct{
	commandData *commandData
	commandType string
	Document *types.Document
}
func NewDeleteCommandData(key string, change_vector []string) (*DeleteCommandData, error){
	ref := &DeleteCommandData{}
	ref.commandData, _ = NewCommandData(key, data.Metadata{}, change_vector, nil)
	ref.commandType = "DELETE"
	return ref, nil
}
func (ref *DeleteCommandData) ToJson() map[string]interface{}{
	out := make(map[string]interface{})
	out["Type"] = ref.commandType
	out["Id"] = ref.commandData.Key
	out["ChangeVector"] = ref.commandData.ChangeVector
	return out
}
func (ref *DeleteCommandData) GetCommand() bool{
	return ref.commandData.GetCommand()
}