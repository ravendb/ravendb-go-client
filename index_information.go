package ravendb

import "time"

type IndexInformation struct {
	Name             string        `json:"Name"`
	IsStale          bool          `json:"IsStale"`
	State            IndexState    `json:"State"`
	LockMode         IndexLockMode `json:"LockMode"`
	Priority         IndexPriority `json:"Priority"`
	Type             IndexType     `json:"Type"`
	LastIndexingTime ServerTime    `json:"LastIndexingTime"`
}

func (i *IndexInformation) GetLastIndexingTime() time.Time {
	return time.Time(i.LastIndexingTime)
}
