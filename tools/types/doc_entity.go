package types

import "github.com/ravendb-go-client/data"

//todo: type is not implement
type TDocByID map[string]*Document
func (ref TDocByID) Clear() {
	ref = make(TDocByID,0)
}
type TDocByEntities map[*interface{}]*Document
func (ref TDocByEntities) Clear() {
	ref = make(TDocByEntities,0)
}
func (ref TDocByEntities) GetKeyByValue(val *Document) (*interface{}, bool){
	for k, v := range ref{
		if v == val{
			return k, true
		}
	}
	return nil, false
}
func (ref TDocByEntities) HasValue(val *Document) bool{
	for _, v := range ref{
		if v == val{
			return true
		}
	}
	return false
}
type Document struct{
	OriginalValue         interface{}
	Metadata              *data.Metadata
	OriginalMetadata      *data.Metadata
	ChangeVector          []string
	Key                   string
	ForceConcurrencyCheck bool
}
