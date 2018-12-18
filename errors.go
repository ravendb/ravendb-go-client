package ravendb

import "fmt"

type errorBase struct {
	ErrorStr string
}

// Error makes it conform to error interface
func (e *errorBase) Error() string {
	return e.ErrorStr
}

type RuntimeError struct {
	errorBase
}

func newRuntimeError(format string, args ...interface{}) *RuntimeError {
	res := &RuntimeError{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

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

type ConcurrencyError struct {
	RavenError

	ExpectedETag         int
	ActualETag           int
	ExpectedChangeVector string
	ActualChangeVector   string
}

func NewConcurrencyError(format string, args ...interface{}) *ConcurrencyError {
	res := &ConcurrencyError{}
	res.RavenError = *newRavenError(format, args...)
	return res
}

type UnsupportedOperationError struct {
	errorBase
}

func newUnsupportedOperationError(format string, args ...interface{}) *UnsupportedOperationError {
	res := &UnsupportedOperationError{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

type IllegalStateError struct {
	errorBase
}

func newIllegalStateError(format string, args ...interface{}) *IllegalStateError {
	res := &IllegalStateError{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

type IllegalArgumentError struct {
	errorBase
}

func newIllegalArgumentError(format string, args ...interface{}) *IllegalArgumentError {
	res := &IllegalArgumentError{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

type NotImplementedError struct {
	errorBase
}

func newNotImplementedError(format string, args ...interface{}) *NotImplementedError {
	res := &NotImplementedError{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// NonUniqueObjectException corresponds to Java's NonUniqueObjectException
type NonUniqueObjectException struct {
	errorBase
}

// NewNonUniqueObjectException creates new NonUniqueObjectError
func NewNonUniqueObjectException(format string, args ...interface{}) *NonUniqueObjectException {
	res := &NonUniqueObjectException{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// DatabaseDoesNotExistException corresponds to Java's DatabaseDoesNotExistException
type DatabaseDoesNotExistException struct {
	errorBase
}

// NewDatabaseDoesNotExistException creates new NonUniqueObjectError
func NewDatabaseDoesNotExistException(format string, args ...interface{}) *DatabaseDoesNotExistException {
	res := &DatabaseDoesNotExistException{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// AllTopologyNodesDownException corresponds to Java's AllTopologyNodesDownException
type AllTopologyNodesDownException struct {
	errorBase
}

// NewAllTopologyNodesDownException creates new AllTopologyNodesDownException
func NewAllTopologyNodesDownException(format string, args ...interface{}) *AllTopologyNodesDownException {
	res := &AllTopologyNodesDownException{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// OperationCancelledException corresponds to Java's OperationCancelledException
type OperationCancelledException struct {
	errorBase
}

// NewOperationCancelledException creates new OperationCancelledException
func NewOperationCancelledException(format string, args ...interface{}) *OperationCancelledException {
	res := &OperationCancelledException{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// AuthorizationException corresponds to Java's AuthorizationException
type AuthorizationException struct {
	errorBase
}

func NewAuthorizationException(format string, args ...interface{}) *AuthorizationException {
	res := &AuthorizationException{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// TimeoutException corresponds to Java's TimeoutException
type TimeoutException struct {
	errorBase
}

func NewTimeoutException(format string, args ...interface{}) *TimeoutException {
	res := &TimeoutException{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// IndexDoesNotExistException corresponds to Java's IndexDoesNotExistException
type IndexDoesNotExistException struct {
	errorBase
}

func NewIndexDoesNotExistException(format string, args ...interface{}) *IndexDoesNotExistException {
	res := &IndexDoesNotExistException{}
	res.errorBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// BadResponseException corresponds to Java's BadResponseException
type BadResponseException struct {
	errorBase
}

func NewBadResponseException(format string, args ...interface{}) *BadResponseException {
	res := &BadResponseException{}
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
