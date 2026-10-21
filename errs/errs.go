// Package errs provides application level error classification shared by services.
//
// Domain and use-case layers return *errs.AppError to describe a failure in
// business terms: a kind, a stable machine code, a safe message and optional
// field level details. Transport layers translate the kind into an HTTP status
// and the code into a client facing string.
//
//	err := errs.NotFound("USER_NOT_FOUND", "user not found")
//	// in a gin handler:
//	ginresp.WriteError(c, err)
//
// The root cause is kept in Cause for logging only and is never returned to
// the client.
//
// This package also owns the Code type and the common codes, so the domain
// layer and the transport layer share one vocabulary. transport/response
// re-exports them as aliases.
package errs

import "errors"

// Code is a stable machine readable business code, e.g. "USER_NOT_FOUND".
// It is the shared vocabulary between the domain layer and the transport
// layer, so a code means the same thing everywhere.
type Code string

// Common business codes. Define more in your service as needed.
const (
	// CodeBadRequest is a generic malformed request.
	CodeBadRequest Code = "BAD_REQUEST"
	// CodeUnauthorized is a missing or invalid authentication.
	CodeUnauthorized Code = "UNAUTHORIZED"
	// CodeForbidden is an authenticated caller without permission.
	CodeForbidden Code = "FORBIDDEN"
	// CodeNotFound is a missing resource.
	CodeNotFound Code = "NOT_FOUND"
	// CodeInternalError is an unexpected failure.
	CodeInternalError Code = "INTERNAL_ERROR"
	// CodeInvalidParam is an invalid input.
	CodeInvalidParam Code = "INVALID_PARAM"
)

// Kind classifies an application error.
type Kind int

// Error kinds.
const (
	// KindInternal is an unexpected failure: a bug or a dependency outage.
	KindInternal Kind = iota
	// KindNotFound means the requested resource does not exist.
	KindNotFound
	// KindValidation means the input failed validation.
	KindValidation
	// KindConflict means the request conflicts with the current state.
	KindConflict
	// KindUnauthorized means the caller is not authenticated.
	KindUnauthorized
	// KindForbidden means the caller is authenticated but not allowed.
	KindForbidden
)

// String returns a stable, human readable name of the kind.
func (kind Kind) String() string {
	switch kind {
	case KindNotFound:
		return "not_found"
	case KindValidation:
		return "validation"
	case KindConflict:
		return "conflict"
	case KindUnauthorized:
		return "unauthorized"
	case KindForbidden:
		return "forbidden"
	case KindInternal:
		return "internal"
	default:
		return "unknown"
	}
}

// FieldError describes a single validation error on a field.
type FieldError struct {
	// Field is the name of the offending input field.
	Field string `json:"field"`
	// Message explains what is wrong with the field.
	Message string `json:"message"`
}

// AppError is an application error that is safe to expose to clients.
type AppError struct {
	// Kind classifies the failure and drives the HTTP status mapping.
	Kind Kind
	// Code is a stable machine readable code, e.g. "USER_NOT_FOUND".
	Code Code
	// Message is a safe, human readable message returned to the client.
	Message string
	// Details carries the optional field level details. It is always a
	// list, so the client can parse it the same way for every error.
	Details []FieldError
	// Cause is the root cause, kept for logging only.
	Cause error
}

// Error implements the error interface.
func (e *AppError) Error() string {
	return e.Message
}

// Unwrap exposes the root cause to errors.Is and errors.As.
func (e *AppError) Unwrap() error {
	return e.Cause
}

// newAppError builds an AppError with every field set.
func newAppError(kind Kind, code Code, message string, details []FieldError, cause error) *AppError {
	return &AppError{
		Kind:    kind,
		Code:    code,
		Message: message,
		Details: details,
		Cause:   cause,
	}
}

// Internal builds an internal error, usually from an unexpected failure.
func Internal(code Code, message string, cause ...error) *AppError {
	return newAppError(KindInternal, code, message, nil, firstCause(cause))
}

// NotFound builds a not found error.
func NotFound(code Code, message string, cause ...error) *AppError {
	return newAppError(KindNotFound, code, message, nil, firstCause(cause))
}

// Validation builds a validation error carrying optional field level details.
func Validation(code Code, message string, details ...FieldError) *AppError {
	return newAppError(KindValidation, code, message, details, nil)
}

// Conflict builds a conflict error.
func Conflict(code Code, message string, cause ...error) *AppError {
	return newAppError(KindConflict, code, message, nil, firstCause(cause))
}

// Unauthorized builds an unauthorized error.
func Unauthorized(code Code, message string, cause ...error) *AppError {
	return newAppError(KindUnauthorized, code, message, nil, firstCause(cause))
}

// Forbidden builds a forbidden error.
func Forbidden(code Code, message string, cause ...error) *AppError {
	return newAppError(KindForbidden, code, message, nil, firstCause(cause))
}

// firstCause returns the first error of the list, or nil.
func firstCause(cause []error) error {
	if len(cause) == 0 {
		return nil
	}
	return cause[0]
}

// From extracts the first *AppError from the error chain.
func From(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}

// HasCode reports whether the error chain carries the given application code.
func HasCode(err error, code Code) bool {
	appErr, ok := From(err)
	return ok && appErr.Code == code
}
