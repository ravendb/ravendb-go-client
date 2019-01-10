package ravendb

// SubscriptionCreationOptions describes options for creating a subscription
type SubscriptionCreationOptions struct {
	name         string
	query        string
	changeVector *string
	mentorNode   string
}
