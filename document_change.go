package ravendb

// DocumentChange describes a change to the document. Can be used as DatabaseChange.
type DocumentChange struct {
	Type           DocumentChangeTypes
	ID             string
	CollectionName string
	ChangeVector   *string
}

func (c *DocumentChange) String() string {
	return c.Type + " on " + c.ID
}
