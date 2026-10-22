// Package ginresp writes the standard response envelope from gin handlers.
//
// The request ID is read from the gin context, so the
// middlewares.RequestID middleware must be registered on the router:
//
//	r.Use(middlewares.RequestID())
//	r.GET("/todos", func(c *gin.Context) {
//		ginresp.OK(c, todos)
//	})
package ginresp

import (
	"net/http"

	"github.com/ductran999/shared-pkg/errs"
	"github.com/ductran999/shared-pkg/transport/response"
	"github.com/gin-gonic/gin"
)

// RequestIDKey is the gin context key holding the request ID.
const RequestIDKey = "request_id"

// RequestIDFrom returns the request ID stored in the context, if any.
func RequestIDFrom(c *gin.Context) string {
	return c.GetString(RequestIDKey)
}

// requestID returns the request ID stored in the context, if any.
func requestID(c *gin.Context) string {
	return RequestIDFrom(c)
}

// OK writes a 200 success envelope.
func OK[T any](c *gin.Context, data T) {
	c.JSON(http.StatusOK, response.OK(data).WithRequestID(requestID(c)))
}

// Created writes a 201 success envelope.
func Created[T any](c *gin.Context, data T) {
	c.JSON(http.StatusCreated, response.Created(data).WithRequestID(requestID(c)))
}

// List writes a 200 success envelope with pagination metadata.
func List[T any](c *gin.Context, items []T, p response.Pagination) {
	c.JSON(http.StatusOK, response.List(items, p).WithRequestID(requestID(c)))
}

// BadRequest writes a 400 envelope with code BAD_REQUEST.
func BadRequest(c *gin.Context, msg ...string) {
	c.JSON(http.StatusBadRequest, response.BadRequest(msg...).WithRequestID(requestID(c)))
}

// BadRequestCode writes a 400 envelope with a custom code and optional field errors.
func BadRequestCode(c *gin.Context, code response.Code, message string, fieldErrors ...response.FieldError) {
	c.JSON(http.StatusBadRequest, response.NewError(code, message, fieldErrors).WithRequestID(requestID(c)))
}

// Unauthorized writes a 401 envelope.
func Unauthorized(c *gin.Context, msg ...string) {
	c.JSON(http.StatusUnauthorized, response.Unauthorized(msg...).WithRequestID(requestID(c)))
}

// Forbidden writes a 403 envelope.
func Forbidden(c *gin.Context, msg ...string) {
	c.JSON(http.StatusForbidden, response.Forbidden(msg...).WithRequestID(requestID(c)))
}

// NotFound writes a 404 envelope.
func NotFound(c *gin.Context, msg ...string) {
	c.JSON(http.StatusNotFound, response.NotFound(msg...).WithRequestID(requestID(c)))
}

// InternalError writes a 500 envelope.
func InternalError(c *gin.Context, msg ...string) {
	c.JSON(http.StatusInternalServerError, response.InternalError(msg...).WithRequestID(requestID(c)))
}

// Error writes a custom code and message with the given HTTP status.
func Error(c *gin.Context, httpStatus int, code response.Code, message string, fieldErrors ...response.FieldError) {
	c.JSON(httpStatus, response.NewError(code, message, fieldErrors).WithRequestID(requestID(c)))
}

// StatusFor maps an application error kind to its HTTP status.
func StatusFor(kind errs.Kind) int {
	switch kind {
	case errs.KindNotFound:
		return http.StatusNotFound
	case errs.KindValidation:
		return http.StatusBadRequest
	case errs.KindConflict:
		return http.StatusConflict
	case errs.KindUnauthorized:
		return http.StatusUnauthorized
	case errs.KindForbidden:
		return http.StatusForbidden
	case errs.KindInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// WriteError maps an application error to the matching HTTP status and writes
// the error envelope. Errors that are not *errs.AppError are treated as
// internal errors, so unexpected failures never leak to the client.
//
//	err := errs.NotFound("USER_NOT_FOUND", "user not found")
//	ginresp.WriteError(c, err)
//	// {"error":{"code":"USER_NOT_FOUND","message":"user not found"},"request_id":"..."}
func WriteError(c *gin.Context, err error) {
	appErr, ok := errs.From(err)
	if !ok {
		InternalError(c, "internal server error")
		return
	}

	envelope := response.NewError(appErr.Code, appErr.Message, appErr.Details)
	c.JSON(StatusFor(appErr.Kind), envelope.WithRequestID(requestID(c)))
}
