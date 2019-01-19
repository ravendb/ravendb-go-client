package ravendb

// SubscriptionState describes state of subscription
type SubscriptionState struct {
	Query                                 string  `json:"Query"`
	ChangeVectorForNextBatchStartingPoint *string `json:"ChangeVectorForNextBatchStartingPoint"`
	SubscriptionID                        int64   `json:"SubscriptionId"`
	SubscriptionName                      string  `json:"SubscriptionName"`
	MentorNode                            string  `json:"MentorNode"`
	NodeTag                               string  `json:"NodeTag"`
	LastBatchAckTime                      Time    `json:"LastBatchAckTime"`
	LastClientConnectionTime              Time    `json:"LastClientConnectionTime"`
	Disabled                              bool    `json:"Disabled"`
}
