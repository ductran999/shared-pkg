// Package ginresp writes the standard response envelope from gin handlers.
//
// Example:
//
//	r.Use(ginresp.RequestID())
//	r.GET("/todos", func(c *gin.Context) {
//		ginresp.OK(c, todos)
//	})
package ginresp

import (
	"net/http"

	"github.com/ductran999/shared-pkg/transport/response"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDKey is the gin context key holding the request ID.
const RequestIDKey = "request_id"

// RequestID generates a UUIDv4 request ID per request, stores it in the gin
// context and echoes it back in the X-Request-ID response header.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(RequestIDKey, uuid.NewString())
		c.Writer.Header().Set("X-Request-ID", c.GetString(RequestIDKey))
		c.Next()
	}
}

// requestID returns the request ID stored in the context, if any.
func requestID(c *gin.Context) string {
	return c.GetString(RequestIDKey)
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

// BadRequestCode writes a 400 envelope with a custom code (e.g. INVALID_PARAM).
func BadRequestCode(c *gin.Context, code response.Code, msg ...string) {
	message := "Bad Request"
	if len(msg) > 0 && msg[0] != "" {
		message = msg[0]
	}
	c.JSON(http.StatusBadRequest, response.NewError(code, message).WithRequestID(requestID(c)))
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
func Error(c *gin.Context, httpStatus int, code response.Code, message string, errs ...response.FieldError) {
	c.JSON(httpStatus, response.NewError(code, message, errs...).WithRequestID(requestID(c)))
}
