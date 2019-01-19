package ravendb

// SubscriptionCreationOptions describes options for creating a subscription
type SubscriptionCreationOptions struct {
	Name         string  `json:"Name"`
	Query        string  `json:"Query"`
	ChangeVector *string `json:"ChangeVector"`
	MentorNode   string  `json:"MentorNode,omitempty"`
}
