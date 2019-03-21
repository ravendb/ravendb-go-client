package ravendb

// subscriptionServerMessageType describes type of subscription server message
type subscriptionServerMessageType = string

const (
	//subscriptionServerMessageNone             = "None"
	subscriptionServerMessageConnectionStatus = "ConnectionStatus"
	subscriptionServerMessageEndOfBatch       = "EndOfBatch"
	subscriptionServerMessageData             = "Data"
	subscriptionServerMessageConfirm          = "Confirm"
	subscriptionServerMessageError            = "Error"
)

// subscriptionConnectionStatus describes subscription connection status
type subscriptionConnectionStatus = string

const (
	//subscriptionConnectionStatusNone                 = "None"
	subscriptionConnectionStatusAccepted = "Accepted"
	subscriptionConnectionStatusInUse    = "InUse"
	subscriptionConnectionStatusClosed   = "Closed"
	subscriptionConnectionStatusNotFound = "NotFound"
	subscriptionConnectionStatusRedirect = "Redirect"
	//subscriptionConnectionStatusForbiddenReadOnly    = "ForbiddenReadOnly"
	//subscriptionConnectionStatusForbidden            = "Forbidden"
	subscriptionConnectionStatusInvalid              = "Invalid"
	subscriptionConnectionStatusConcurrencyReconnect = "ConcurrencyReconnect"
)

// subscriptionRedirectData describes subscription redirect data
/*
type subscriptionRedirectData struct {
	currentTag    string
	redirectedTag string
}
*/

// subscriptionConnectionServerMessage describes subscription connection server message
type subscriptionConnectionServerMessage struct {
	Type      subscriptionServerMessageType `json:"Type"`
	Status    subscriptionConnectionStatus  `json:"Status"`
	Data      map[string]interface{}        `json:"Data"`
	Exception string                        `json:"Exception"`
	Message   string                        `json:"Message"`
}
