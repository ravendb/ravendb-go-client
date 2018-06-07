package ravendb

type OperationStatusChange struct {
	operationId int
	state       ObjectNode
}

/*
   public long getOperationId() {
       return operationId;
   }

   public void setOperationId(long operationId) {
       this.operationId = operationId;
   }

   public ObjectNode getState() {
       return state;
   }

   public void setState(ObjectNode state) {
       this.state = state;
   }
*/
