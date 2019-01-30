package ravendb

// ScriptResolver describes script resolver
type ScriptResolver struct {
	Script           string `json:"Script"`
	LastModifiedTime Time   `json:"LastModifiedTime"`
}
