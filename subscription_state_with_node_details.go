package ravendb

// SubscriptionStateWithNodeDetails describes subscription state with node details
type SubscriptionStateWithNodeDetails struct {
	SubscriptionState
	ResponsibleNode NodeID
}
