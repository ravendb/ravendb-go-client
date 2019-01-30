package ravendb

// PutResult describes result of PutDocumentCommand
type PutResult struct {
	ID           string  `json:"Id"`
	ChangeVector *string `json:"ChangeVector"`
}
