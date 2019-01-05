package ravendb

// DatabaseTopology describes a topology of the database
type DatabaseTopology struct {
	Members                  []string          `json:"Members"`
	Promotables              []string          `json:"Promotables"`
	Rehabs                   []string          `json:"Rehabs"`
	PredefinedMentors        map[string]string `json:"PredefinedMentors"`
	DemotionReasons          map[string]string `json:"DemotionReasons"`
	PromotablesStatus        map[string]string `json:"PromotablesStatus"`
	ReplicationFactor        int               `json:"ReplicationFactor"`
	DynamicNodesDistribution bool              `json:"DynamicNodesDistribution"`
	Stamp                    LeaderStamp       `json:"Stamp"`
}
