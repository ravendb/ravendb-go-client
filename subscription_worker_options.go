package ravendb

import "time"

// SubscriptionWorkerOptions describes options for subscription worker
type SubscriptionWorkerOptions struct {
	SubscriptionName                string                      `json:"SubscriptionName"`
	TimeToWaitBeforeConnectionRetry Duration                    `json:"TimeToWaitBeforeConnectionRetry"`
	IgnoreSubscriberErrors          bool                        `json:"IgnoreSubscriberErrors"`
	Strategy                        SubscriptionOpeningStrategy `json:"Strategy"`
	MaxDocsPerBatch                 int                         `json:"MaxDocsPerBatch"`
	MaxErroneousPeriod              Duration                    `json:"MaxErroneousPeriod"`
	CloseWhenNoDocsLeft             bool                        `json:"CloseWhenNoDocsLeft"`
}

// NewSubscriptionWorkerOptions returns new SubscriptionWorkerOptions
func NewSubscriptionWorkerOptions(subscriptionName string) *SubscriptionWorkerOptions {
	panicIf(subscriptionName == "", "SubscriptionName cannot be null or empty")
	return &SubscriptionWorkerOptions{
		Strategy:                        SubscriptionOpeningStrategyOpenIfFree,
		MaxDocsPerBatch:                 4096,
		TimeToWaitBeforeConnectionRetry: Duration(time.Second * 5),
		MaxErroneousPeriod:              Duration(time.Minute * 5),
		SubscriptionName:                subscriptionName,
	}
}
