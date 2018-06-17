package ravendb

import "fmt"

type ExceptionBase struct {
	ErrorStr string
}

// Error makes it conform to error interface
func (e *ExceptionBase) Error() string {
	return e.ErrorStr
}

type RuntimeException struct {
	ExceptionBase
}

func NewRuntimeException(format string, args ...interface{}) *RuntimeException {
	res := &RuntimeException{}
	res.ExceptionBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

type RavenException struct {
	ExceptionBase
}

func NewRavenException(format string, args ...interface{}) *RavenException {
	res := &RavenException{}
	res.ExceptionBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

type ConflictException struct {
	RavenException
}

func NewConflictException(format string, args ...interface{}) *ConflictException {
	res := &ConflictException{}
	res.RavenException = *NewRavenException(format, args)
	return res
}

type ConcurrencyException struct {
	RavenException

	ExpectedETag         int
	ActualETag           int
	ExpectedChangeVector string
	ActualChangeVector   string
}

func NewConcurrencyException(format string, args ...interface{}) *ConcurrencyException {
	res := &ConcurrencyException{}
	res.RavenException = *NewRavenException(format, args)
	return res
}

type UnsupportedOperationException struct {
	ExceptionBase
}

func NewUnsupportedOperationException(format string, args ...interface{}) *UnsupportedOperationException {
	res := &UnsupportedOperationException{}
	res.ExceptionBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// IllegalStateException corresponds to Java's IllegalStateException
type IllegalStateException struct {
	ExceptionBase
}

// NewIllegalStateException creates a new IllegalStateError
func NewIllegalStateException(format string, args ...interface{}) *IllegalStateException {
	res := &IllegalStateException{}
	res.ExceptionBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// IllegalArgumentException corresponds to Java's IllegalArgumentException
type IllegalArgumentException struct {
	ExceptionBase
}

// NewIllegalArgumentException creates new IllegalArgumentError
func NewIllegalArgumentException(format string, args ...interface{}) *IllegalArgumentException {
	res := &IllegalArgumentException{}
	res.ExceptionBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// NotImplementedException corresponds to Java's NotImplementedException
type NotImplementedException struct {
	ExceptionBase
}

// NewNotImplementedException creates new NotImplementedError
func NewNotImplementedException(format string, args ...interface{}) *NotImplementedException {
	res := &NotImplementedException{}
	res.ExceptionBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// NonUniqueObjectException corresponds to Java's NonUniqueObjectException
type NonUniqueObjectException struct {
	ExceptionBase
}

// NewNonUniqueObjectException creates new NonUniqueObjectError
func NewNonUniqueObjectException(format string, args ...interface{}) *NonUniqueObjectException {
	res := &NonUniqueObjectException{}
	res.ExceptionBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// DatabaseDoesNotExistException corresponds to Java's DatabaseDoesNotExistException
type DatabaseDoesNotExistException struct {
	ExceptionBase
}

// NewDatabaseDoesNotExistException creates new NonUniqueObjectError
func NewDatabaseDoesNotExistException(format string, args ...interface{}) *DatabaseDoesNotExistException {
	res := &DatabaseDoesNotExistException{}
	res.ExceptionBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// AllTopologyNodesDownException corresponds to Java's AllTopologyNodesDownException
type AllTopologyNodesDownException struct {
	ExceptionBase
}

// NewAllTopologyNodesDownException creates new AllTopologyNodesDownException
func NewAllTopologyNodesDownException(format string, args ...interface{}) *AllTopologyNodesDownException {
	res := &AllTopologyNodesDownException{}
	res.ExceptionBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// OperationCancelledException corresponds to Java's OperationCancelledException
type OperationCancelledException struct {
	ExceptionBase
}

// NewOperationCancelledException creates new OperationCancelledException
func NewOperationCancelledException(format string, args ...interface{}) *OperationCancelledException {
	res := &OperationCancelledException{}
	res.ExceptionBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// AuthorizationException corresponds to Java's AuthorizationException
type AuthorizationException struct {
	ExceptionBase
}

func NewAuthorizationException(format string, args ...interface{}) *AuthorizationException {
	res := &AuthorizationException{}
	res.ExceptionBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// TimeoutException corresponds to Java's TimeoutException
type TimeoutException struct {
	ExceptionBase
}

func NewTimeoutException(format string, args ...interface{}) *TimeoutException {
	res := &TimeoutException{}
	res.ExceptionBase.ErrorStr = fmt.Sprintf(format, args...)
	return res
}

// BadResponseException corresponds to Java's BadResponseException
type BadResponseException struct {
	ExceptionBase
}

func NewBadResponseException(format string, args ...interface{}) *BadResponseException {
	res := &BadResponseException{}
	res.ExceptionBase.ErrorStr = fmt.Sprintf(format, args...)
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
