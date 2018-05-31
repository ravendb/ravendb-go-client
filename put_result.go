package ravendb

// PutResult describes result of PutDocumentCommand
type PutResult struct {
	ID           string `json:"Id"`
	ChangeVector string `json:"ChangeVector"`
}

/*
public String getId() {
	return id;
}

public void setId(String id) {
	this.id = id;
}

public String getChangeVector() {
	return changeVector;
}

public void setChangeVector(String changeVector) {
	this.changeVector = changeVector;
}
*/
