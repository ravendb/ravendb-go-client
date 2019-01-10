package ravendb

// SubscriptionOpeningStrategy describes opening strategy for subscriptions
type SubscriptionOpeningStrategy = string

const (
	// SubscriptionOpeningStrategyOpenIfFree:
	// The client will successfully open a subscription only if there isn't any other currently connected client.
	// Otherwise it will end up with SubscriptionInUseError
	SubscriptionOpeningStrategyOpenIfFree = "OpenIfFree"
	// SubscriptionOpeningStrategyTakeOver:
	// The connecting client will successfully open a subscription even if there is another active subscription's consumer.
	// If the new client takes over an existing client then the existing one will get a SubscriptionInUseException.
	//
	// The subscription will always be held by the last connected client.
	SubscriptionOpeningStrategyTakeOver = "TakeOver"
	// SubscriptionOpeningStrategyWaitForFree:
	// If the client currently cannot open the subscription because it is used by another client but it will wait for that client
	// to complete and keep attempting to gain the subscription
	SubscriptionOpeningStrategyWaitForFree = "WaitForFree"
)
