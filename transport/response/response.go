// Package response defines the standard API envelopes returned by
// services: Success for happy paths and Error for failures.
//
// Example:
//
//	return response.OK(todos)
//	return response.List(items, response.NewPagination(page, size, total))
//	return response.BadRequest("invalid id")
//	return response.NewError("OBJECT_ALREADY_EXISTED", "object already existed", nil)
package response

import "github.com/ductran999/shared-pkg/errs"

// Code is a custom business code. It is an alias of errs.Code, so the domain
// layer and the transport layer share one vocabulary and one set of constants.
type Code = errs.Code

// Common business codes.
const (
	CodeBadRequest    = errs.CodeBadRequest
	CodeUnauthorized  = errs.CodeUnauthorized
	CodeForbidden     = errs.CodeForbidden
	CodeNotFound      = errs.CodeNotFound
	CodeInternalError = errs.CodeInternalError
	CodeInvalidParam  = errs.CodeInvalidParam
)

// FieldError describes a single validation error.
// It is an alias of errs.FieldError so the domain layer can build
// validation details without importing the transport package.
type FieldError = errs.FieldError

// Success is the success envelope.
//
//	{"data":{...},"meta":{...},"request_id":"..."}
type Success[T any] struct {
	Data      T      `json:"data"`
	Meta      any    `json:"meta,omitempty"`
	RequestID string `json:"request_id"`
}

// WithRequestID returns a copy of the success envelope carrying the request ID.
func (s Success[T]) WithRequestID(requestID string) Success[T] {
	s.RequestID = requestID
	return s
}

// ErrorDetail carries the error information.
type ErrorDetail struct {
	Code    Code         `json:"code"`
	Message string       `json:"message"`
	Details []FieldError `json:"details,omitempty"`
}

// Error is the error envelope.
//
//	{"error":{"code":"...","message":"...","details":{...}},"request_id":"..."}
type Error struct {
	Error     ErrorDetail `json:"error"`
	RequestID string      `json:"request_id"`
}

// WithRequestID returns a copy of the error envelope carrying the request ID.
func (e Error) WithRequestID(requestID string) Error {
	e.RequestID = requestID
	return e
}

// OK returns a success envelope with data.
func OK[T any](data T) Success[T] {
	return Success[T]{Data: data}
}

// Created returns a success envelope with the created resource.
func Created[T any](data T) Success[T] {
	return Success[T]{Data: data}
}

// List returns a success envelope with items and pagination metadata.
func List[T any](items []T, p Pagination) Success[[]T] {
	return Success[[]T]{Data: items, Meta: p}
}

// NewError builds an error envelope with a custom code, message and optional details.
func NewError(code Code, message string, details []FieldError) Error {
	return Error{Error: ErrorDetail{Code: code, Message: message, Details: details}}
}

// BadRequest returns a BAD_REQUEST envelope. The message defaults to "Bad Request".
func BadRequest(msg ...string) Error {
	return NewError(CodeBadRequest, firstOrMessage(msg, "Bad Request"), nil)
}

// Unauthorized returns an UNAUTHORIZED envelope. The message defaults to "Unauthorized".
func Unauthorized(msg ...string) Error {
	return NewError(CodeUnauthorized, firstOrMessage(msg, "Unauthorized"), nil)
}

// Forbidden returns a FORBIDDEN envelope. The message defaults to "Forbidden".
func Forbidden(msg ...string) Error {
	return NewError(CodeForbidden, firstOrMessage(msg, "Forbidden"), nil)
}

// NotFound returns a NOT_FOUND envelope. The message defaults to "Not Found".
func NotFound(msg ...string) Error {
	return NewError(CodeNotFound, firstOrMessage(msg, "Not Found"), nil)
}

// InternalError returns an INTERNAL_ERROR envelope. The message defaults to "Internal Server Error".
func InternalError(msg ...string) Error {
	return NewError(CodeInternalError, firstOrMessage(msg, "Internal Server Error"), nil)
}

// firstOrMessage returns the first non-empty message, or fallback.
func firstOrMessage(values []string, fallback string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return fallback
}
