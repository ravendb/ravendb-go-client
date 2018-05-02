package types

//todo: type is not implement
type TDocByID map[string]string

func (ref TDocByID) Clear() {
	ref = make(TDocByID, 0)
}

type TDocByEntities map[string]*TDocByEntity

func (ref TDocByEntities) Clear() {
	ref = make(TDocByEntities, 0)
}

type TDocByEntity struct {
	Original_value          string
	Metadata                string
	Original_metadata       string
	Change_vector           string
	Key                     string
	Force_concurrency_check bool
}
