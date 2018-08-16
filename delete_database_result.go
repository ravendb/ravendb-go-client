package ravendb

// DeleteDatabaseResult represents result of Delete Database command
type DeleteDatabaseResult struct {
	RaftCommandIndex int      `json:"RaftCommandIndex"`
	PendingDeletes   []string `json:"PendingDeletes"`
}

/*
public long getRaftCommandIndex() {
	return raftCommandIndex;
}

public void setRaftCommandIndex(long raftCommandIndex) {
	this.raftCommandIndex = raftCommandIndex;
}

public string[] getPendingDeletes() {
	return pendingDeletes;
}

public void setPendingDeletes(string[] pendingDeletes) {
	this.pendingDeletes = pendingDeletes;
}
*/
