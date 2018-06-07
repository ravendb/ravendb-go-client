package ravendb

type DocumentChange struct {
	typ DocumentChangeTypes

	id String

	collectionName String

	typeName String

	changeVector *String
}

/*
    public DocumentChangeTypes getType() {
        return type;
    }

	 public void setType(DocumentChangeTypes type) {
        this.type = type;
    }

    public String getId() {
        return id;
    }

    public void setId(String id) {
        this.id = id;
    }

    public String getCollectionName() {
        return collectionName;
    }

    public void setCollectionName(String collectionName) {
        this.collectionName = collectionName;
    }

    public String getTypeName() {
        return typeName;
    }

    public void setTypeName(String typeName) {
        this.typeName = typeName;
    }

    public String getChangeVector() {
        return changeVector;
    }

    public void setChangeVector(String changeVector) {
        this.changeVector = changeVector;
    }
*/
