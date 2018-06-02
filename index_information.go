package ravendb

import "time"

type IndexInformation struct {
	Name             String        `json:"Name"`
	Stale            bool          `json:"IsStale"`
	State            IndexState    `json:"State"`
	LockMode         IndexLockMode `json:"LockMode"`
	Priority         IndexPriority `json:"Priority"`
	Type             IndexType     `json:"Type"`
	LastIndexingTime time.Time     `json:"LastIndexingTime"` // TODO: custom marshaller
}

/*
    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public boolean isStale() {
        return stale;
    }

    public void setStale(boolean stale) {
        this.stale = stale;
    }

    public IndexState getState() {
        return state;
    }

    public void setState(IndexState state) {
        this.state = state;
    }

    public IndexLockMode getLockMode() {
        return lockMode;
    }

    public void setLockMode(IndexLockMode lockMode) {
        this.lockMode = lockMode;
    }

    public IndexPriority getPriority() {
        return priority;
    }

    public void setPriority(IndexPriority priority) {
        this.priority = priority;
    }

    public IndexType getType() {
        return type;
    }

    public void setType(IndexType type) {
        this.type = type;
    }

    public Date getLastIndexingTime() {
        return lastIndexingTime;
    }

    public void setLastIndexingTime(Date lastIndexingTime) {
        this.lastIndexingTime = lastIndexingTime;
    }
}
*/
