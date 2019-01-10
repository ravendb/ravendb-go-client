package ravendb

// GetSubscriptionsResult represents result of "get subscriptions"
type GetSubscriptionsResult struct {
	Results []SubscriptionState `json:"Results"`
}
