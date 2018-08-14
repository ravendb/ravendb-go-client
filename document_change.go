package ravendb

type DocumentChange struct {
	typ DocumentChangeTypes

	id string

	collectionName string

	typeName string

	changeVector *string
}

/*
    public DocumentChangeTypes getType() {
        return type;
    }

	 public void setType(DocumentChangeTypes type) {
        this.type = type;
    }

    public string GetId() {
        return id;
    }

    public void setId(string id) {
        this.id = id;
    }

    public string getCollectionName() {
        return collectionName;
    }

    public void setCollectionName(string collectionName) {
        this.collectionName = collectionName;
    }

    public string getTypeName() {
        return typeName;
    }

    public void setTypeName(string typeName) {
        this.typeName = typeName;
    }

    public string getChangeVector() {
        return changeVector;
    }

    public void setChangeVector(string changeVector) {
        this.changeVector = changeVector;
    }
*/
