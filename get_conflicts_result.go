package ravendb

type GetConflictsResult struct {
	ID          string      `json:"Id"`
	Results     []*Conflict `json:"Results"`
	LargestEtag int         `json:"LargestEtag"`
}

/*
   public String GetId() {
       return id;
   }

   public void setId(String id) {
       this.id = id;
   }

   public Conflict[] getResults() {
       return results;
   }

   public void setResults(Conflict[] results) {
       this.results = results;
   }

   public long getLargestEtag() {
       return largestEtag;
   }

   public void setLargestEtag(long largestEtag) {
       this.largestEtag = largestEtag;
   }
*/

type Conflict struct {
	LastModified ServerTime `json:"LastModified"`
	ChangeVector string     `json:"ChangeVector"`
	Doc          ObjectNode `json:"Doc"`
}

/*
       public Date getLastModified() {
           return lastModified;
       }

       public void setLastModified(Date lastModified) {
           this.lastModified = lastModified;
       }

       public String getChangeVector() {
           return changeVector;
       }

       public void setChangeVector(String changeVector) {
           this.changeVector = changeVector;
       }

       public ObjectNode getDoc() {
           return doc;
       }

       public void setDoc(ObjectNode doc) {
           this.doc = doc;
       }
   }
*/
