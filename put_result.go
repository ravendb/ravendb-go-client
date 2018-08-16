package ravendb

// PutResult describes result of PutDocumentCommand
type PutResult struct {
	ID           string  `json:"Id"`
	ChangeVector *string `json:"ChangeVector"`
}

func (r *PutResult) GetID() string {
	return r.ID
}

func (r *PutResult) GetChangeVector() *string {
	return r.ChangeVector
}

/*
public void setId(string id) {
	this.id = id;
}


public void setChangeVector(string changeVector) {
	this.changeVector = changeVector;
}
*/
