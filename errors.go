package ravendb

import (
	"fmt"
	"strings"
)

type CancellationError struct {
}

func (e *CancellationError) Error() string {
	return "CancellationError"
}

type errorBase struct {
	wrapped  error
	ErrorStr string
}

// Error makes it conform to error interface
func (e *errorBase) Error() string {
	return e.ErrorStr
}

// hackish way to get a wrapped error
func (e *errorBase) WrappedError() error {
	return e.wrapped
}

type iWrappedError interface {
	WrappedError() error
}

// GetWrappedError returns an error wrapped by this error
// If no error is wrapped, returns nil
func GetWrappedError(err error) error {
	if e, ok := err.(iWrappedError); ok {
		return e.WrappedError()
	}
	return nil
}

func (e *errorBase) setErrorf(format string, args ...interface{}) {
	if len(args) == 0 {
		e.ErrorStr = format
		return
	}
	// a bit of a hack: to make it easy to port Java code, if the last
	// argument is of type error, we consider it a wrapped error
	n := len(args)
	last := args[n-1]
	if err, ok := last.(error); ok {
		e.wrapped = err
		args = args[:n-1]
	}
	e.ErrorStr = fmt.Sprintf(format, args...)
}

// RuntimeError represents generic runtime error
type RuntimeError struct {
	errorBase
}

func newRuntimeError(format string, args ...interface{}) *RuntimeError {
	res := &RuntimeError{}
	res.setErrorf(format, args...)
	return res
}

// UnsupportedOperationError represents unsupported operation error
type UnsupportedOperationError struct {
	errorBase
}

func newUnsupportedOperationError(format string, args ...interface{}) *UnsupportedOperationError {
	res := &UnsupportedOperationError{}
	res.setErrorf(format, args...)
	return res
}

// IllegalStateError represents illegal state error
type IllegalStateError struct {
	errorBase
}

func newIllegalStateError(format string, args ...interface{}) *IllegalStateError {
	res := &IllegalStateError{}
	res.setErrorf(format, args...)
	return res
}

// IllegalArgumentError represents illegal argument error
type IllegalArgumentError struct {
	errorBase
}

func newIllegalArgumentError(format string, args ...interface{}) *IllegalArgumentError {
	res := &IllegalArgumentError{}
	res.setErrorf(format, args...)
	return res
}

// NotImplementedError represents not implemented error
type NotImplementedError struct {
	errorBase
}

func newNotImplementedError(format string, args ...interface{}) *NotImplementedError {
	res := &NotImplementedError{}
	res.setErrorf(format, args...)
	return res
}

// AllTopologyNodesDownError represents "all topology nodes are down" error
type AllTopologyNodesDownError struct {
	errorBase
}

func newAllTopologyNodesDownError(format string, args ...interface{}) *AllTopologyNodesDownError {
	res := &AllTopologyNodesDownError{}
	res.setErrorf(format, args...)
	return res
}

// OperationCancelledError represents "operation cancelled" error
type OperationCancelledError struct {
	errorBase
}

func newOperationCancelledError(format string, args ...interface{}) *OperationCancelledError {
	res := &OperationCancelledError{}
	res.setErrorf(format, args...)
	return res
}

// AuthorizationError represents authorization error
type AuthorizationError struct {
	errorBase
}

func newAuthorizationError(format string, args ...interface{}) *AuthorizationError {
	res := &AuthorizationError{}
	res.setErrorf(format, args...)
	return res
}

// RavenError represents generic raven error
// all exceptions that in Java extend RavenException should
// contain this error
type RavenError struct {
	errorBase
}

// hackish way to see if "inherits" from (embeds) RavenError
func (e *RavenError) isRavenError() bool {
	return true
}

type iRavenError interface {
	isRavenError() bool
}

func isRavenError(err error) bool {
	_, ok := err.(iRavenError)
	return ok
}

func newRavenError(format string, args ...interface{}) *RavenError {
	res := &RavenError{}
	res.setErrorf(format, args...)
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
	res.setErrorf(format, args...)
	return res
}

// NonUniqueObjectError represents non unique object error
type NonUniqueObjectError struct {
	RavenError
}

// newNonUniqueObjectError creates new NonUniqueObjectError
func newNonUniqueObjectError(format string, args ...interface{}) *NonUniqueObjectError {
	res := &NonUniqueObjectError{}
	res.setErrorf(format, args...)
	return res
}

// DatabaseDoesNotExistError represents "database not not exist" error
type DatabaseDoesNotExistError struct {
	RavenError
}

// newDatabaseDoesNotExistError creates new NonUniqueObjectError
func newDatabaseDoesNotExistError(format string, args ...interface{}) *DatabaseDoesNotExistError {
	res := &DatabaseDoesNotExistError{}
	res.setErrorf(format, args...)
	return res
}

// TimeoutError represents timeout error
type TimeoutError struct {
	RavenError
}

// NewTimeoutError returns new TimeoutError
func NewTimeoutError(format string, args ...interface{}) *TimeoutError {
	res := &TimeoutError{}
	res.setErrorf(format, args...)
	return res
}

// IndexDoesNotExistError represents "index doesn't exist" error
type IndexDoesNotExistError struct {
	RavenError
}

func newIndexDoesNotExistError(format string, args ...interface{}) *IndexDoesNotExistError {
	res := &IndexDoesNotExistError{}
	res.setErrorf(format, args...)
	return res
}

// BadResponseError represents "bad response" error
type BadResponseError struct {
	RavenError
}

func newBadResponseError(format string, args ...interface{}) *BadResponseError {
	res := &BadResponseError{}
	res.setErrorf(format, args...)
	return res
}

// BadRequestError maps to server's 400 Bad Request response
// This is additional information sent by the server
type BadRequestError struct {
	RavenError
}

// ConflictError maps to server's 409 Conflict response
type ConflictError struct {
	RavenError
}

func newConflictError(format string, args ...interface{}) *ConflictError {
	res := &ConflictError{}
	res.setErrorf(format, args...)
	return res
}

// a base type for subscription-related errors
type SubscriptionError struct {
	RavenError
}

// SubscriberErrorError represents error about subscriber error
// Note: name is unfortunate but it corresponds to Java's SubscriberErrorException
type SubscriberErrorError struct {
	SubscriptionError
}

// SubscriptionChangeVectorUpdateConcurrencyError represents an error about
// subscription change vector update concurrency
type SubscriptionChangeVectorUpdateConcurrencyError struct {
	SubscriptionError
}

func newSubscriptionChangeVectorUpdateConcurrencyError(format string, args ...interface{}) *SubscriptionChangeVectorUpdateConcurrencyError {
	res := &SubscriptionChangeVectorUpdateConcurrencyError{}
	res.setErrorf(format, args...)
	return res
}

// SubscriptionClosedError is returned when subscription is closed
type SubscriptionClosedError struct {
	SubscriptionError
}

func newSubscriptionClosedError(format string, args ...interface{}) *SubscriptionClosedError {
	res := &SubscriptionClosedError{}
	res.setErrorf(format, args...)
	return res
}

// SubscriptionDoesNotBelongToNodeError is returned when subscription
// does not belong to node
type SubscriptionDoesNotBelongToNodeError struct {
	SubscriptionError

	appropriateNode string
}

func newSubscriptionDoesNotBelongToNodeError(format string, args ...interface{}) *SubscriptionDoesNotBelongToNodeError {
	res := &SubscriptionDoesNotBelongToNodeError{}
	res.setErrorf(format, args...)
	return res
}

// SubscriptionDoesNotExistError is returned when subscription doesn't exist
type SubscriptionDoesNotExistError struct {
	SubscriptionError
}

func newSubscriptionDoesNotExistError(format string, args ...interface{}) *SubscriptionDoesNotExistError {
	res := &SubscriptionDoesNotExistError{}
	res.setErrorf(format, args...)
	return res
}

// SubscriptionInvalidStateError is returned when subscription is in invalid state
type SubscriptionInvalidStateError struct {
	SubscriptionError
}

func newSubscriptionInvalidStateError(format string, args ...interface{}) *SubscriptionInvalidStateError {
	res := &SubscriptionInvalidStateError{}
	res.setErrorf(format, args...)
	return res
}

// SubscriptionInUseError is returned when subscription is in use
type SubscriptionInUseError struct {
	SubscriptionError
}

func newSubscriptionInUseError(format string, args ...interface{}) *SubscriptionInUseError {
	res := &SubscriptionInUseError{}
	res.setErrorf(format, args...)
	return res
}

// ClientVersionMismatchError is returned when subscription is in use
type ClientVersionMismatchError struct {
	RavenError
}

// CertificateNameMismatchError is returned when subscription is in use
type CertificateNameMismatchError struct {
	errorBase
}

func throwCancellationRequested() error {
	return newOperationCancelledError("")
}

type InvalidQueryError struct {
	RavenError
}

type UnsuccessfulRequestError struct {
	RavenError
}

type ChangeProcessingError struct {
	RavenError
}

type CommandExecutionError struct {
	RavenError
}

type NodeIsPassiveError struct {
	RavenError
}

type NoLeaderError struct {
	RavenError
}

type LicenseActivationError struct {
	RavenError
}

type CompilationError struct {
	RavenError
}

type DatabaseConcurrentLoadTimeoutError struct {
	RavenError
}

type DatabaseDisabledError struct {
	RavenError
}

type DatabaseLoadFailureError struct {
	RavenError
}

type DatabaseLoadTimeoutError struct {
	RavenError
}

type DatabaseNotRelevantError struct {
	RavenError
}

type DocumentDoesNotExistError struct {
	RavenError
}

type BulkInsertProtocolViolationError struct {
	RavenError
}

type IndexAlreadyExistError struct {
	RavenError
}

type IndexCreationError struct {
	RavenError
}

type IndexDeletionError struct {
	RavenError
}

type IndexInvalidError struct {
	RavenError
}

type JavaScriptError struct {
	RavenError
}

type RevisionsDisabledError struct {
	RavenError
}

type RouteNotFoundError struct {
	RavenError
}

type SecurityError struct {
	RavenError
}

type ServerLoadFailureError struct {
	RavenError
}

type IndexCompilationError struct {
	RavenError
}

func makeRavenErrorFromName(exceptionName string, errMsg string) error {
	// Java's "FooException" is "FooError" in Go
	s := strings.Replace(exceptionName, "Exception", "Error", -1)
	switch s {
	case "IndexCompilationError":
		res := &IndexCompilationError{}
		res.ErrorStr = errMsg
		return res
	case "ConcurrencyError":
		res := &ConcurrencyError{}
		res.ErrorStr = errMsg
		return res
	case "NonUniqueObjectError":
		res := &NonUniqueObjectError{}
		res.ErrorStr = errMsg
		return res
	case "DatabaseDoesNotExistError":
		res := &DatabaseDoesNotExistError{}
		res.ErrorStr = errMsg
		return res
	case "TimeoutError":
		res := &TimeoutError{}
		res.ErrorStr = errMsg
		return res
	case "IndexDoesNotExistError":
		res := &IndexDoesNotExistError{}
		res.ErrorStr = errMsg
		return res
	case "BadResponseError":
		res := &BadResponseError{}
		res.ErrorStr = errMsg
		return res
	case "BadRequestError":
		res := &BadRequestError{}
		res.ErrorStr = errMsg
		return res
	case "ConflictError":
		res := &ConflictError{}
		res.ErrorStr = errMsg
		return res
	case "SubscriberErrorError":
		res := &SubscriberErrorError{}
		res.ErrorStr = errMsg
		return res
	case "SubscriptionChangeVectorUpdateConcurrencyError":
		res := &SubscriptionChangeVectorUpdateConcurrencyError{}
		res.ErrorStr = errMsg
		return res
	case "SubscriptionClosedError":
		res := &SubscriptionClosedError{}
		res.ErrorStr = errMsg
		return res
	case "SubscriptionDoesNotBelongToNodeError":
		res := &SubscriptionDoesNotBelongToNodeError{}
		res.ErrorStr = errMsg
		return res
	case "SubscriptionDoesNotExistError":
		res := &SubscriptionDoesNotExistError{}
		res.ErrorStr = errMsg
		return res
	case "SubscriptionInvalidStateError":
		res := &SubscriptionInvalidStateError{}
		res.ErrorStr = errMsg
		return res
	case "SubscriptionInUseError":
		res := &SubscriptionInUseError{}
		res.ErrorStr = errMsg
		return res
	case "ClientVersionMismatchError":
		res := &ClientVersionMismatchError{}
		res.ErrorStr = errMsg
		return res
	case "CertificateNameMismatchError":
		res := &CertificateNameMismatchError{}
		res.ErrorStr = errMsg
		return res
	case "InvalidQueryError":
		res := &InvalidQueryError{}
		res.ErrorStr = errMsg
		return res
	case "UnsuccessfulRequestError":
		res := &UnsuccessfulRequestError{}
		res.ErrorStr = errMsg
		return res
	case "ChangeProcessingError":
		res := &ChangeProcessingError{}
		res.ErrorStr = errMsg
		return res
	case "CommandExecutionError":
		res := &CommandExecutionError{}
		res.ErrorStr = errMsg
		return res
	case "NodeIsPassiveError":
		res := &NodeIsPassiveError{}
		res.ErrorStr = errMsg
		return res
	case "NoLeaderError":
		res := &NoLeaderError{}
		res.ErrorStr = errMsg
		return res
	case "LicenseActivationError":
		res := &LicenseActivationError{}
		res.ErrorStr = errMsg
		return res
	case "CompilationError":
		res := &CompilationError{}
		res.ErrorStr = errMsg
		return res
	case "DatabaseConcurrentLoadTimeoutError":
		res := &DatabaseConcurrentLoadTimeoutError{}
		res.ErrorStr = errMsg
		return res
	case "DatabaseDisabledError":
		res := &DatabaseDisabledError{}
		res.ErrorStr = errMsg
		return res
	case "DatabaseLoadFailureError":
		res := &DatabaseLoadFailureError{}
		res.ErrorStr = errMsg
		return res
	case "DatabaseLoadTimeoutError":
		res := &DatabaseLoadTimeoutError{}
		res.ErrorStr = errMsg
		return res
	case "DatabaseNotRelevantError":
		res := &DatabaseNotRelevantError{}
		res.ErrorStr = errMsg
		return res
	case "DocumentDoesNotExistError":
		res := &DocumentDoesNotExistError{}
		res.ErrorStr = errMsg
		return res
	case "BulkInsertAbortedError":
		res := &BulkInsertAbortedError{}
		res.ErrorStr = errMsg
		return res
	case "BulkInsertProtocolViolationError":
		res := &BulkInsertProtocolViolationError{}
		res.ErrorStr = errMsg
		return res
	case "IndexAlreadyExistError":
		res := &IndexAlreadyExistError{}
		res.ErrorStr = errMsg
		return res
	case "IndexCreationError":
		res := &IndexCreationError{}
		res.ErrorStr = errMsg
		return res
	case "IndexDeletionError":
		res := &IndexDeletionError{}
		res.ErrorStr = errMsg
		return res
	case "IndexInvalidError":
		res := &IndexInvalidError{}
		res.ErrorStr = errMsg
		return res
	case "JavaScriptError":
		res := &JavaScriptError{}
		res.ErrorStr = errMsg
		return res
	case "RevisionsDisabledError":
		res := &RevisionsDisabledError{}
		res.ErrorStr = errMsg
		return res
	case "RouteNotFoundError":
		res := &RouteNotFoundError{}
		res.ErrorStr = errMsg
		return res
	case "SecurityError":
		res := &SecurityError{}
		res.ErrorStr = errMsg
		return res
	case "ServerLoadFailureError":
		res := &ServerLoadFailureError{}
		res.ErrorStr = errMsg
		return res

	}
	return nil
}
