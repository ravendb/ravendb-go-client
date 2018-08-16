package ravendb

import "time"

type IndexInformation struct {
	Name             string        `json:"Name"`
	Stale            bool          `json:"IsStale"`
	State            IndexState    `json:"State"`
	LockMode         IndexLockMode `json:"LockMode"`
	Priority         IndexPriority `json:"Priority"`
	Type             IndexType     `json:"Type"`
	LastIndexingTime ServerTime    `json:"LastIndexingTime"`
}

func (i *IndexInformation) GetName() string {
	return i.Name
}

func (i *IndexInformation) IsStale() bool {
	return i.Stale
}

func (i *IndexInformation) GetState() IndexState {
	return i.State
}

func (i *IndexInformation) GetLockMode() IndexLockMode {
	return i.LockMode
}
func (i *IndexInformation) GetPriority() IndexPriority {
	return i.Priority
}

func (i *IndexInformation) GetType() IndexType {
	return i.Type
}

func (i *IndexInformation) GetLastIndexingTime() time.Time {
	return time.Time(i.LastIndexingTime)
}

/*
    public void setName(string name) {
        this.name = name;
    }

    public void setStale(boolean stale) {
        this.stale = stale;
    }

    public void setState(IndexState state) {
        this.state = state;
    }

    public void setLockMode(IndexLockMode lockMode) {
        this.lockMode = lockMode;
    }

    public void setPriority(IndexPriority priority) {
        this.priority = priority;
    }

    public void setType(IndexType type) {
        this.type = type;
    }

    public void setLastIndexingTime(Date lastIndexingTime) {
        this.lastIndexingTime = lastIndexingTime;
    }
}
*/
