package ravendb

// PutResult describes result of PutDocumentCommand
type PutResult struct {
	ID           string `json:"Id"`
	ChangeVector string `json:"ChangeVector"`
}

/*
public string getId() {
	return id;
}

public void setId(string id) {
	this.id = id;
}

public string getChangeVector() {
	return changeVector;
}

public void setChangeVector(string changeVector) {
	this.changeVector = changeVector;
}
*/
