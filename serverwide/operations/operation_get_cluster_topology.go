package operations

import (
	"encoding/json"
	"github.com/ravendb/ravendb-go-client"
	"net/http"
)

type OperationGetClusterTopology struct {
	Topology struct {
		TopologyId  string            `json:"TopologyId"`
		AllNodes    map[string]string `json:"AllNodes"`
		Members     map[string]string `json:"Members"`
		Promotables map[string]string `json:"Promotables"`
		Watchers    map[string]string `json:"Watchers"`
		LastNodeId  string            `json:"LastNodeId"`
		Etag        int               `json:"Etag"`
	} `json:"Topology"`
	Etag               int    `json:"Etag"`
	Leader             string `json:"Leader"`
	LeaderShipDuration int    `json:"LeaderShipDuration"`
	CurrentState       string `json:"CurrentState"`
	NodeTag            string `json:"NodeTag"`
	CurrentTerm        int    `json:"CurrentTerm"`
	NodeLicenseDetails struct {
		UtilizedCores       int         `json:"UtilizedCores"`
		MaxUtilizedCores    interface{} `json:"MaxUtilizedCores"`
		NumberOfCores       int         `json:"NumberOfCores"`
		InstalledMemoryInGb float64     `json:"InstalledMemoryInGb"`
		UsableMemoryInGb    float64     `json:"UsableMemoryInGb"`
		BuildInfo           struct {
			ProductVersion string `json:"ProductVersion"`
			BuildVersion   int    `json:"BuildVersion"`
			CommitHash     string `json:"CommitHash"`
			FullVersion    string `json:"FullVersion"`
		} `json:"BuildInfo"`
		OsInfo struct {
			Type         string `json:"Type"`
			FullName     string `json:"FullName"`
			Version      string `json:"Version"`
			BuildVersion string `json:"BuildVersion"`
			Is64Bit      bool   `json:"Is64Bit"`
		} `json:"OsInfo"`
	}
	LastStateChangeReason string   `json:"LastStateChangeReason"`
	Status                struct{} `json:"Status"`
}

func (operation *OperationGetClusterTopology) GetCommand(conventions *ravendb.DocumentConventions) (ravendb.RavenCommand, error) {
	return &getClusterTopology{
		RavenCommandBase: ravendb.RavenCommandBase{
			ResponseType: ravendb.RavenCommandResponseTypeObject,
		},
		parent: operation,
	}, nil
}

type getClusterTopology struct {
	ravendb.RavenCommandBase
	parent *OperationGetClusterTopology
}

func (c *getClusterTopology) CreateRequest(node *ravendb.ServerNode) (*http.Request, error) {
	url := node.URL + "/cluster/topology"
	return http.NewRequest(http.MethodGet, url, nil)
}
func (c *getClusterTopology) SetResponse(response []byte, fromCache bool) error {
	return json.Unmarshal(response, c.parent)
}
