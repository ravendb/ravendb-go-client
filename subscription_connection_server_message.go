package ravendb

// SubscriptionServerMessageType describes type of subscription server message
type SubscriptionServerMessageType = string

const (
	SubscriptionServerMessageNone             = "None"
	SubscriptionServerMessageConnectionStatus = "ConnectionStatus"
	SubscriptionServerMessageEndOfBatch       = "EndOfBatch"
	SubscriptionServerMessageData             = "Data"
	SubscriptionServerMessageConfirm          = "Confirm"
	SubscriptionServerMessageError            = "Error"
)

// SubscriptionConnectionStatus describes subscription connection status
type SubscriptionConnectionStatus = string

const (
	SubscriptionConnectionStatusNone                 = "None"
	SubscriptionConnectionStatusAccepted             = "Accepted"
	SubscriptionConnectionStatusInUse                = "InUse"
	SubscriptionConnectionStatusClosed               = "Closed"
	SubscriptionConnectionStatusNotFound             = "NotFound"
	SubscriptionConnectionStatusRedirect             = "Redirect"
	SubscriptionConnectionStatusForbiddenReadOnly    = "ForbiddenReadOnly"
	SubscriptionConnectionStatusForbidden            = "Forbidden"
	SubscriptionConnectionStatusInvalid              = "Invalid"
	SubscriptionConnectionStatusConcurrencyReconnect = "ConcurrencyReconnect"
)

// SubscriptionRedirectData describes subscription redirect data
// TODO: make private?
type SubscriptionRedirectData struct {
	currentTag    string
	redirectedTag string
}

// SubscriptionConnectionServerMessage describes subscription connection server message
// TODO: make private?
type SubscriptionConnectionServerMessage struct {
	Type      SubscriptionServerMessageType `json:"Type"`
	Status    SubscriptionConnectionStatus  `json:"Status"`
	Data      map[string]interface{}        `json:"Data"`
	Exception string                        `json:"Exception"`
	Message   string                        `json:"Message"`
}
