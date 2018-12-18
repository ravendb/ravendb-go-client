package ravendb

type scriptResolver struct {
	Script           string `json:"Script"`
	LastModifiedTime Time   `json:"LastModifiedTime"`
}
