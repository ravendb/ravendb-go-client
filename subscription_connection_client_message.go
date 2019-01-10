package ravendb

// SubscriptionClientMessageType describes type of subscription client message
type SubscriptionClientMessageType = string

const (
	SubscriptionClientMessageNone                 = "None"
	SubscriptionClientMessageAcknowledge          = "Acknowledge"
	SubscriptionClientMessageDisposedNotification = "DisposedNotification"
)

// SubscriptionConnectionClientMessage describes a subscription connection message
type SubscriptionConnectionClientMessage struct {
	Type         SubscriptionClientMessageType `json:"Type"`
	ChangeVector *string                       `json:"ChangeVector"`
}
