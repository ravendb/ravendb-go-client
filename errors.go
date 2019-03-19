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

type LicenseActivation struct {
	RavenError
}

func makeRavenErrorFromName(s string) error {
	// Java's "FooException" is "FooError" in Go
	s = strings.Replace(s, "Exception", "Error", -1)
	switch s {
	case "LicenseActivation":
		return &LicenseActivation{}
	case "ConcurrencyError":
		return &ConcurrencyError{}
	case "NonUniqueObjectError":
		return &NonUniqueObjectError{}
	case "DatabaseDoesNotExistError":
		return &DatabaseDoesNotExistError{}
	case "TimeoutError":
		return &TimeoutError{}
	case "IndexDoesNotExistError":
		return &IndexDoesNotExistError{}
	case "BadResponseError":
		return &BadResponseError{}
	case "BadRequestError":
		return &BadRequestError{}
	case "ConflictError":
		return &ConflictError{}
	case "SubscriberErrorError":
		return &SubscriberErrorError{}
	case "SubscriptionChangeVectorUpdateConcurrencyError":
		return &SubscriptionChangeVectorUpdateConcurrencyError{}
	case "SubscriptionClosedError":
		return &SubscriptionClosedError{}
	case "SubscriptionDoesNotBelongToNodeError":
		return &SubscriptionDoesNotBelongToNodeError{}
	case "SubscriptionDoesNotExistError":
		return &SubscriptionDoesNotExistError{}
	case "SubscriptionInvalidStateError":
		return &SubscriptionInvalidStateError{}
	case "SubscriptionInUseError":
		return &SubscriptionInUseError{}
	case "ClientVersionMismatchError":
		return &ClientVersionMismatchError{}
	case "CertificateNameMismatchError":
		return &CertificateNameMismatchError{}
	case "InvalidQueryError":
		return &InvalidQueryError{}
	case "UnsuccessfulRequestError":
		return &UnsuccessfulRequestError{}
	case "ChangeProcessingError":
		return &ChangeProcessingError{}
	case "CommandExecutionError":
		return &CommandExecutionError{}
	case "NodeIsPassiveError":
		return &NodeIsPassiveError{}
	case "NoLeaderError":
		return &NoLeaderError{}
	case "LicenseActivationError":
		return &LicenseActivationError{}
	case "CompilationError":
		return &CompilationError{}
	case "DatabaseConcurrentLoadTimeoutError":
		return &DatabaseConcurrentLoadTimeoutError{}
	case "DatabaseDisabledError":
		return &DatabaseDisabledError{}
	case "DatabaseLoadFailureError":
		return &DatabaseLoadFailureError{}
	case "DatabaseLoadTimeoutError":
		return &DatabaseLoadTimeoutError{}
	case "DatabaseNotRelevantError":
		return &DatabaseNotRelevantError{}
	case "DocumentDoesNotExistError":
		return &DocumentDoesNotExistError{}
	case "BulkInsertAbortedError":
		return &BulkInsertAbortedError{}
	case "BulkInsertProtocolViolationError":
		return &BulkInsertProtocolViolationError{}
	case "IndexAlreadyExistError":
		return &IndexAlreadyExistError{}
	case "IndexCreationError":
		return &IndexCreationError{}
	case "IndexDeletionError":
		return &IndexDeletionError{}
	case "IndexInvalidError":
		return &IndexInvalidError{}
	case "JavaScriptError":
		return &JavaScriptError{}
	case "RevisionsDisabledError":
		return &RevisionsDisabledError{}
	case "RouteNotFoundError":
		return &RouteNotFoundError{}
	case "SecurityError":
		return &SecurityError{}
	case "ServerLoadFailureError":
		return &ServerLoadFailureError{}

	}
	return nil
}
