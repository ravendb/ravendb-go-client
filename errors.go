package ravendb

import "fmt"

type errorBase struct {
	ErrorStr string
}

// Error makes it conform to error interface
func (e *errorBase) Error() string {
	return e.ErrorStr
}

// RuntimeError represents generic runtime error
type RuntimeError struct {
	errorBase
}

func newRuntimeError(format string, args ...interface{}) *RuntimeError {
	res := &RuntimeError{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// RavenError represents generic raven error
type RavenError struct {
	errorBase
}

func newRavenError(format string, args ...interface{}) *RavenError {
	res := &RavenError{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
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

	ExpectedETag         int
	ActualETag           int
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
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// IllegalStateError represents illegal state error
type IllegalStateError struct {
	errorBase
}

func newIllegalStateError(format string, args ...interface{}) *IllegalStateError {
	res := &IllegalStateError{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// IllegalArgumentError represents illegal argument error
type IllegalArgumentError struct {
	errorBase
}

func newIllegalArgumentError(format string, args ...interface{}) *IllegalArgumentError {
	res := &IllegalArgumentError{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// NotImplementedError represents not implemented error
type NotImplementedError struct {
	errorBase
}

func newNotImplementedError(format string, args ...interface{}) *NotImplementedError {
	res := &NotImplementedError{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// NonUniqueObjectError represents non unique object error
type NonUniqueObjectError struct {
	errorBase
}

// newNonUniqueObjectError creates new NonUniqueObjectError
func newNonUniqueObjectError(format string, args ...interface{}) *NonUniqueObjectError {
	res := &NonUniqueObjectError{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// DatabaseDoesNotExistError represents "database not not exist" error
type DatabaseDoesNotExistError struct {
	errorBase
}

// newDatabaseDoesNotExistError creates new NonUniqueObjectError
func newDatabaseDoesNotExistError(format string, args ...interface{}) *DatabaseDoesNotExistError {
	res := &DatabaseDoesNotExistError{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// AllTopologyNodesDownError represents "all topology nodes are down" error
type AllTopologyNodesDownError struct {
	errorBase
}

func newAllTopologyNodesDownError(format string, args ...interface{}) *AllTopologyNodesDownError {
	res := &AllTopologyNodesDownError{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// OperationCancelledError represents "operation cancelled" error
type OperationCancelledError struct {
	errorBase
}

func newOperationCancelledError(format string, args ...interface{}) *OperationCancelledError {
	res := &OperationCancelledError{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// AuthorizationError represents authorization error
type AuthorizationError struct {
	errorBase
}

func newAuthorizationError(format string, args ...interface{}) *AuthorizationError {
	res := &AuthorizationError{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// TimeoutError represents timeout error
type TimeoutError struct {
	errorBase
}

// NewTimeoutError returns new TimeoutError
func NewTimeoutError(format string, args ...interface{}) *TimeoutError {
	res := &TimeoutError{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// IndexDoesNotExistError represents "index doesn't exist" error
type IndexDoesNotExistError struct {
	errorBase
}

func newIndexDoesNotExistError(format string, args ...interface{}) *IndexDoesNotExistError {
	res := &IndexDoesNotExistError{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// BadResponseError represents "bad response" error
type BadResponseError struct {
	errorBase
}

func newBadResponseError(format string, args ...interface{}) *BadResponseError {
	res := &BadResponseError{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
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

func NewCancellationError() *CancellationError {
	return &CancellationError{}
}
