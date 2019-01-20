package ravendb

// SubscriptionCreationOptions describes options for creating a subscription
type SubscriptionCreationOptions struct {
	// must omitempty Name or else the server will try to find a subscription
	// with empty name
	Name         string  `json:"Name,omitempty"`
	Query        string  `json:"Query"`
	ChangeVector *string `json:"ChangeVector"`
	MentorNode   string  `json:"MentorNode,omitempty"`
}
