package ravendb

type PatchResult struct {
	Status           PatchStatus `json:"Status"`
	ModifiedDocument ObjectNode  `json:"ModifiedDocument"`
	OriginalDocument ObjectNode  `json:"OriginalDocument"`
	Debug            ObjectNode  `json:"Debug"`

	// TODO: can this ever be null? If not, use string for type
	ChangeVector *string `json:"ChangeVector"`
	Collection   string  `json:"Collection"`
}

func (r *PatchResult) getStatus() PatchStatus {
	return r.Status
}

func (r *PatchResult) getModifiedDocument() ObjectNode {
	return r.ModifiedDocument
}

func (r *PatchResult) getOriginalDocument() ObjectNode {
	return r.OriginalDocument
}

func (r *PatchResult) getDebug() ObjectNode {
	return r.Debug
}

func (r *PatchResult) getChangeVector() *string {
	return r.ChangeVector
}

func (r *PatchResult) getCollection() string {
	return r.Collection
}

/*
   public void setStatus(PatchStatus status) {
       this.status = status;
   }


   public void setModifiedDocument(ObjectNode modifiedDocument) {
       this.modifiedDocument = modifiedDocument;
   }

   public void setOriginalDocument(ObjectNode originalDocument) {
       this.originalDocument = originalDocument;
   }

   public void setDebug(ObjectNode debug) {
       this.debug = debug;
   }

   public void setChangeVector(String changeVector) {
       this.changeVector = changeVector;
   }


   public void setCollection(String collection) {
       this.collection = collection;
   }
*/
