package ravendb

import "fmt"

type errorBase struct {
	ErrorStr string
}

// Error makes it conform to error interface
func (e *errorBase) Error() string {
	return e.ErrorStr
}

func (e *errorBase) setErrorf(format string, args ...interface{}) {
	if len(args) == 0 {
		e.ErrorStr = format
		return
	}
	e.ErrorStr = fmt.Sprintf(format, args...)
}

// RuntimeError represents generic runtime error
type RuntimeError struct {
	errorBase
}

func newRuntimeError(format string, args ...interface{}) *RuntimeError {
	res := &RuntimeError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

// RavenError represents generic raven error
type RavenError struct {
	errorBase
}

func newRavenError(format string, args ...interface{}) *RavenError {
	res := &RavenError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

type ConflictException struct {
	RavenError
}

func NewConflictException(format string, args ...interface{}) *ConflictException {
	res := &ConflictException{}
	res.RavenError = *newRavenError(format, args...)
	return res
}

// ConcurrencyError represents concurrency error
type ConcurrencyError struct {
	RavenError

	ExpectedETag         int64
	ActualETag           int64
	ExpectedChangeVector string
	ActualChangeVector   string
}

func newConcurrencyError(format string, args ...interface{}) *ConcurrencyError {
	res := &ConcurrencyError{}
	res.RavenError = *newRavenError(format, args...)
	return res
}

// UnsupportedOperationError represents unsupported operation error
type UnsupportedOperationError struct {
	errorBase
}

func newUnsupportedOperationError(format string, args ...interface{}) *UnsupportedOperationError {
	res := &UnsupportedOperationError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

// IllegalStateError represents illegal state error
type IllegalStateError struct {
	errorBase
}

func newIllegalStateError(format string, args ...interface{}) *IllegalStateError {
	res := &IllegalStateError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

// IllegalArgumentError represents illegal argument error
type IllegalArgumentError struct {
	errorBase
}

func newIllegalArgumentError(format string, args ...interface{}) *IllegalArgumentError {
	res := &IllegalArgumentError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

// NotImplementedError represents not implemented error
type NotImplementedError struct {
	errorBase
}

func newNotImplementedError(format string, args ...interface{}) *NotImplementedError {
	res := &NotImplementedError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

// NonUniqueObjectError represents non unique object error
type NonUniqueObjectError struct {
	errorBase
}

// newNonUniqueObjectError creates new NonUniqueObjectError
func newNonUniqueObjectError(format string, args ...interface{}) *NonUniqueObjectError {
	res := &NonUniqueObjectError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

// DatabaseDoesNotExistError represents "database not not exist" error
type DatabaseDoesNotExistError struct {
	errorBase
}

// newDatabaseDoesNotExistError creates new NonUniqueObjectError
func newDatabaseDoesNotExistError(format string, args ...interface{}) *DatabaseDoesNotExistError {
	res := &DatabaseDoesNotExistError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

// AllTopologyNodesDownError represents "all topology nodes are down" error
type AllTopologyNodesDownError struct {
	errorBase
}

func newAllTopologyNodesDownError(format string, args ...interface{}) *AllTopologyNodesDownError {
	res := &AllTopologyNodesDownError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

// OperationCancelledError represents "operation cancelled" error
type OperationCancelledError struct {
	errorBase
}

func newOperationCancelledError(format string, args ...interface{}) *OperationCancelledError {
	res := &OperationCancelledError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

// AuthorizationError represents authorization error
type AuthorizationError struct {
	errorBase
}

func newAuthorizationError(format string, args ...interface{}) *AuthorizationError {
	res := &AuthorizationError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

// TimeoutError represents timeout error
type TimeoutError struct {
	errorBase
}

// NewTimeoutError returns new TimeoutError
func NewTimeoutError(format string, args ...interface{}) *TimeoutError {
	res := &TimeoutError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

// IndexDoesNotExistError represents "index doesn't exist" error
type IndexDoesNotExistError struct {
	errorBase
}

func newIndexDoesNotExistError(format string, args ...interface{}) *IndexDoesNotExistError {
	res := &IndexDoesNotExistError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

// BadResponseError represents "bad response" error
type BadResponseError struct {
	errorBase
}

func newBadResponseError(format string, args ...interface{}) *BadResponseError {
	res := &BadResponseError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

// BadRequestError maps to server's 400 Bad Request response
// This is additional information sent by the server
type BadRequestError struct {
	URL      string `json:"Url"`
	Type     string `json:"Type"`
	Message  string `json:"Message"`
	ErrorStr string `json:"Error"`
}

// Error makes it conform to error interface
func (e *BadRequestError) Error() string {
	return fmt.Sprintf(`Server returned 400 Bad Request for URL '%s'
Type: %s
Message: %s
Error: %s`, e.URL, e.Type, e.Message, e.ErrorStr)
}

// InternalServerError maps to server's 500 Internal Server response
type InternalServerError struct {
	URL      string `json:"Url"`
	Type     string `json:"Type"`
	Message  string `json:"Message"`
	ErrorStr string `json:"Error"`
}

// Error makes it conform to error interface
func (e *InternalServerError) Error() string {
	return fmt.Sprintf(`Server returned 500 Internal Server for URL '%s'
Type: %s
Message: %s
Error: %s`, e.URL, e.Type, e.Message, e.ErrorStr)
}

// ServiceUnavailableError maps to server's 501 Service Unavailable
// response. This is additional information sent by the server.
type ServiceUnavailableError struct {
	Type    string `json:"Type"`
	Message string `json:"Message"`
}

// Error makes it conform to error interface
func (e *ServiceUnavailableError) Error() string {
	return fmt.Sprintf(`Server returned 501 Service Unavailable'
Type: %s
Message: %s`, e.Type, e.Message)
}

// ConflictError maps to server's 409 Conflict response
type ConflictError struct {
	URL      string `json:"Url"`
	Type     string `json:"Type"`
	Message  string `json:"Message"`
	ErrorStr string `json:"Error"`
}

// Error makes it conform to error interface
func (e *ConflictError) Error() string {
	return fmt.Sprintf(`Server returned 409 Conflict for URL '%s'
Type: %s
Message: %s
Error: %s`, e.URL, e.Type, e.Message, e.ErrorStr)
}

// NotFoundError maps to server's 404 Not Found
type NotFoundError struct {
	URL string
}

// Error makes it conform to error interface
func (e *NotFoundError) Error() string {
	return fmt.Sprintf(`Server returned 404 Not Found for URL '%s'`, e.URL)
}

type CancellationError struct {
}

func (e *CancellationError) Error() string {
	return "CancellationError"
}

// SubscriberErrorError represents error about subscriber error
// Note: name is unfortunate but it corresponds to Java's SubscriberErrorException
type SubscriberErrorError struct {
	errorBase
}

func newSubscriberErrorError(format string, args ...interface{}) *SubscriberErrorError {
	res := &SubscriberErrorError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

// SubscriptionChangeVectorUpdateConcurrencyError represents an error about
// subscription change vector update concurrency
type SubscriptionChangeVectorUpdateConcurrencyError struct {
	errorBase
}

func newSubscriptionChangeVectorUpdateConcurrencyError(format string, args ...interface{}) *SubscriptionChangeVectorUpdateConcurrencyError {
	res := &SubscriptionChangeVectorUpdateConcurrencyError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

// SubscriptionClosedError is returned when subscription is closed
type SubscriptionClosedError struct {
	errorBase
}

func newSubscriptionClosedError(format string, args ...interface{}) *SubscriptionClosedError {
	res := &SubscriptionClosedError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

// SubscriptionDoesNotBelongToNodeError is returned when subscription does not belong
// to node
type SubscriptionDoesNotBelongToNodeError struct {
	errorBase

	appropriateNode string
}

func newSubscriptionDoesNotBelongToNodeError(format string, args ...interface{}) *SubscriptionDoesNotBelongToNodeError {
	res := &SubscriptionDoesNotBelongToNodeError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

// SubscriptionDoesNotExistError is returned when subscription doesn't exist
type SubscriptionDoesNotExistError struct {
	errorBase
}

func newSubscriptionDoesNotExistError(format string, args ...interface{}) *SubscriptionDoesNotExistError {
	res := &SubscriptionDoesNotExistError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

// SubscriptionError is a generic error related to subscription
type SubscriptionError struct {
	errorBase
}

func newSubscriptionError(format string, args ...interface{}) *SubscriptionError {
	res := &SubscriptionError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

// SubscriptionInvalidStateError is returned when subscription is in invalid state
type SubscriptionInvalidStateError struct {
	errorBase
}

func newSubscriptionInvalidStateError(format string, args ...interface{}) *SubscriptionInvalidStateError {
	res := &SubscriptionInvalidStateError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

// SubscriptionInUseError is returned when subscription is in use
type SubscriptionInUseError struct {
	errorBase
}

func newSubscriptionInUseError(format string, args ...interface{}) *SubscriptionInUseError {
	res := &SubscriptionInUseError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

// ClientVersionMismatchError is returned when subscription is in use
type ClientVersionMismatchError struct {
	errorBase
}

func newClientVersionMismatchError(format string, args ...interface{}) *ClientVersionMismatchError {
	res := &ClientVersionMismatchError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

// CertificateNameMismatchError is returned when subscription is in use
type CertificateNameMismatchError struct {
	errorBase
}

func newCertificateNameMismatchError(format string, args ...interface{}) *CertificateNameMismatchError {
	res := &CertificateNameMismatchError{}
	res.errorBase.setErrorf(format, args...)
	return res
}

func throwCancellationRequested() error {
	return newOperationCancelledError("")
}