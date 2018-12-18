package ravendb

// OperationIDResult is a result of commands like CompactDatabaseCommand
type OperationIDResult struct {
	OperationId int `json:"OperationId"`
}

func (r *OperationIDResult) getOperationId() int {
	return r.OperationId
}

/*
    public void setOperationId(long operationID) {
        this.operationID = operationID;
	}
*/
