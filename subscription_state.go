package ravendb

// SubscriptionState describes state of subscription
// TODO: if serialized, make fields public and json-annotate
type SubscriptionState struct {
	query                                 string
	changeVectorForNextBatchStartingPoint string
	subscriptionId                        int64
	subscriptionName                      string
	mentorNode                            string
	nodeTag                               string
	lastBatchAckTime                      Time
	lastClientConnectionTime              Time
	disabled                              bool
}
