// Package response defines the standard API envelopes returned by
// services: Success for happy paths and Error for failures.
//
// Example:
//
//	return response.OK(todos)
//	return response.List(items, response.Pagination{Page: page, Size: size, Total: total})
//	return response.BadRequest("invalid id")
//	return response.NewError("OBJECT_ALREADY_EXISTED", "object already existed")
package response

// Code is a custom business code. Define more in your service as needed.
type Code string

// Common business codes.
const (
	CodeBadRequest    Code = "BAD_REQUEST"
	CodeUnauthorized  Code = "UNAUTHORIZED"
	CodeForbidden     Code = "FORBIDDEN"
	CodeNotFound      Code = "NOT_FOUND"
	CodeInternalError Code = "INTERNAL_ERROR"
	CodeInvalidParam  Code = "INVALID_PARAM"
)

// FieldError describes a single validation error.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Success is the success envelope.
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

// Error is the error envelope.
type Error struct {
	Code      Code         `json:"code"`
	Message   string       `json:"message"`
	Errors    []FieldError `json:"errors,omitempty"`
	RequestID string       `json:"request_id"`
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

// NewError builds an error envelope with a custom code and message.
func NewError(code Code, message string, errs ...FieldError) Error {
	return Error{Code: code, Message: message, Errors: errs}
}

// BadRequest returns a BAD_REQUEST envelope. The message defaults to "Bad Request".
func BadRequest(msg ...string) Error {
	return Error{Code: CodeBadRequest, Message: firstOrMessage(msg, "Bad Request")}
}

// Unauthorized returns an UNAUTHORIZED envelope. The message defaults to "Unauthorized".
func Unauthorized(msg ...string) Error {
	return Error{Code: CodeUnauthorized, Message: firstOrMessage(msg, "Unauthorized")}
}

// Forbidden returns a FORBIDDEN envelope. The message defaults to "Forbidden".
func Forbidden(msg ...string) Error {
	return Error{Code: CodeForbidden, Message: firstOrMessage(msg, "Forbidden")}
}

// NotFound returns a NOT_FOUND envelope. The message defaults to "Not Found".
func NotFound(msg ...string) Error {
	return Error{Code: CodeNotFound, Message: firstOrMessage(msg, "Not Found")}
}

// InternalError returns an INTERNAL_ERROR envelope. The message defaults to "Internal Server Error".
func InternalError(msg ...string) Error {
	return Error{Code: CodeInternalError, Message: firstOrMessage(msg, "Internal Server Error")}
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
