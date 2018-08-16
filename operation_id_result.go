package ravendb

// OperationIdResult is a result of commands like CompactDatabaseCommand
type OperationIdResult struct {
	OperationId int `json:"OperationId"`
}

func (r *OperationIdResult) getOperationId() int {
	return r.OperationId
}

/*
    public void setOperationId(long operationId) {
        this.operationId = operationId;
	}
*/
