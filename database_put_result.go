package ravendb

// DatabasePutResult decribes server response for e.g. CreateDatabaseCommand
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/serverwide/operations/DatabasePutResult.java#L7
type DatabasePutResult struct {
	RaftCommandIndex int      `json:"RaftCommandIndex"`
	Name             string   `json:"Name"`
	DatabaseTopology Topology `json:"Topology"`
	NodesAddedTo     []string `json:"NodesAddedTo"`
}
