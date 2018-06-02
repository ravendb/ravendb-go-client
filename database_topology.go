package ravendb

// DatabaseTopology describes a topology of the database
// https://sourcegraph.com/github.com/ravendb/ravendb-jvm-client@v4.0/-/blob/src/main/java/net/ravendb/client/serverwide/DatabaseTopology.java#L8
type DatabaseTopology struct {
	Members                  []string          `json:"Members"`
	Promotables              []string          `json:"Promotables"`
	Rehabs                   []string          `json:"Rehabs"`
	PredefinedMentors        map[string]string `json:"PredefinedMentors"` // TODO: not present in JSON response from python test
	DemotionReasons          map[string]string `json:"DemotionReasons"`
	PromotablesStatus        map[string]string `json:"PromotablesStatus"`
	ReplicationFactor        int               `json:"ReplicationFactor"`
	DynamicNodesDistribution bool              `json:"DynamicNodesDistribution"`
	Stamp                    LeaderStamp       `json:"Stamp"`
}
