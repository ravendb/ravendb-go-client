package ravendb

type IndexChange struct {
	typ  IndexChangeTypes
	name String
}

/*
   public IndexChangeTypes getType() {
       return type;
   }

   public void setType(IndexChangeTypes type) {
       this.type = type;
   }

   public String getName() {
       return name;
   }

   public void setName(String name) {
       this.name = name;
   }
*/
