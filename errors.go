package ravendb

import "fmt"

type IllegalStateError struct {
	ErrorStr string
}

func NewIllegalStateError(s string) *IllegalStateError {
	return &IllegalStateError{
		ErrorStr: s,
	}
}

// Error makes it conform to error interface
func (e *IllegalStateError) Error() string {
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
