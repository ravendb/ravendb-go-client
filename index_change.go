package ravendb

type IndexChange struct {
	typ  IndexChangeTypes
	name string
}

/*
   public IndexChangeTypes getType() {
       return type;
   }

   public void setType(IndexChangeTypes type) {
       this.type = type;
   }

   public string GetName() {
       return name;
   }

   public void setName(string name) {
       this.name = name;
   }
*/
