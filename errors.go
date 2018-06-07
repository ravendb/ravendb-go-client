package ravendb

import "fmt"

type UnsupportedOperationException struct {
	ErrorStr string
}

func NewUnsupportedOperationException(format string, args ...interface{}) *UnsupportedOperationException {
	return &UnsupportedOperationException{
		ErrorStr: fmt.Sprintf(format, args...),
	}
}

// Error makes it conform to error interface
func (e *UnsupportedOperationException) Error() string {
	return e.ErrorStr
}

// IllegalStateException corresponds to Java's IllegalStateException
type IllegalStateException struct {
	ErrorStr string
}

// NewIllegalStateException creates a new IllegalStateError
func NewIllegalStateException(format string, args ...interface{}) *IllegalStateException {
	return &IllegalStateException{
		ErrorStr: fmt.Sprintf(format, args...),
	}
}

// Error makes it conform to error interface
func (e *IllegalStateException) Error() string {
	return e.ErrorStr
}

// IllegalArgumentException corresponds to Java's IllegalArgumentException
type IllegalArgumentException struct {
	ErrorStr string
}

// NewIllegalArgumentException creates new IllegalArgumentError
func NewIllegalArgumentException(format string, args ...interface{}) *IllegalArgumentException {
	return &IllegalArgumentException{
		ErrorStr: fmt.Sprintf(format, args...),
	}
}

// Error makes it conform to error interface
func (e *IllegalArgumentException) Error() string {
	return e.ErrorStr
}

// NotImplementedException corresponds to Java's NotImplementedException
type NotImplementedException struct {
	ErrorStr string
}

// NewNotImplementedException creates new NotImplementedError
func NewNotImplementedException(format string, args ...interface{}) *NotImplementedException {
	return &NotImplementedException{
		ErrorStr: fmt.Sprintf(format, args...),
	}
}

// Error makes it conform to error interface
func (e *NotImplementedException) Error() string {
	return e.ErrorStr
}

// NonUniqueObjectException corresponds to Java's NonUniqueObjectException
type NonUniqueObjectException struct {
	ErrorStr string
}

// NewNonUniqueObjectException creates new NonUniqueObjectError
func NewNonUniqueObjectException(format string, args ...interface{}) *NonUniqueObjectException {
	return &NonUniqueObjectException{
		ErrorStr: fmt.Sprintf(format, args...),
	}
}

// Error makes it conform to error interface
func (e *NonUniqueObjectException) Error() string {
	return e.ErrorStr
}

// DatabaseDoesNotExistException corresponds to Java's DatabaseDoesNotExistException
type DatabaseDoesNotExistException struct {
	ErrorStr string
}

// NewDatabaseDoesNotExistException creates new NonUniqueObjectError
func NewDatabaseDoesNotExistException(format string, args ...interface{}) *DatabaseDoesNotExistException {
	return &DatabaseDoesNotExistException{
		ErrorStr: fmt.Sprintf(format, args...),
	}
}

// Error makes it conform to error interface
func (e *DatabaseDoesNotExistException) Error() string {
	return e.ErrorStr
}

// AllTopologyNodesDownException corresponds to Java's AllTopologyNodesDownException
type AllTopologyNodesDownException struct {
	ErrorStr string
}

// NewAllTopologyNodesDownException creates new AllTopologyNodesDownException
func NewAllTopologyNodesDownException(format string, args ...interface{}) *AllTopologyNodesDownException {
	return &AllTopologyNodesDownException{
		ErrorStr: fmt.Sprintf(format, args...),
	}
}

// Error makes it conform to error interface
func (e *AllTopologyNodesDownException) Error() string {
	return e.ErrorStr
}

// OperationCancelledException corresponds to Java's OperationCancelledException
type OperationCancelledException struct {
	ErrorStr string
}

// NewOperationCancelledException creates new OperationCancelledException
func NewOperationCancelledException(format string, args ...interface{}) *OperationCancelledException {
	return &OperationCancelledException{
		ErrorStr: fmt.Sprintf(format, args...),
	}
}

// Error makes it conform to error interface
func (e *OperationCancelledException) Error() string {
	return e.ErrorStr
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
