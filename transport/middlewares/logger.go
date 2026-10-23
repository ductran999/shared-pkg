package middlewares

import (
	"errors"
	"fmt"
	"net/http"
	"runtime"
	"time"

	"github.com/ductran999/shared-pkg/transport/ginresp"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Logger writes a structured access log for every request.
//
// It replaces gin.Logger, which writes plain text that is hard to ship to a
// log aggregator. The log line is JSON and carries the request ID, so a
// request can be traced from the access log to the application logs.
//
// Register it after the request ID middleware:
//
//	r.Use(middlewares.RequestID())
//	r.Use(middlewares.Logger())
//
// The optional skip paths are not logged at all, e.g. health checks.
//
//	middlewares.Logger("/healthz", "/metrics")
func Logger(skipPaths ...string) gin.HandlerFunc {
	skipped := make(map[string]struct{}, len(skipPaths))
	for _, path := range skipPaths {
		skipped[path] = struct{}{}
	}

	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		if _, skip := skipped[path]; skip {
			return
		}

		status := c.Writer.Status()

		event := log.Info()
		switch {
		case status >= http.StatusInternalServerError:
			event = log.Error()
		case status >= http.StatusBadRequest:
			event = log.Warn()
		}

		event = event.
			Str("request_id", ginresp.RequestIDFrom(c)).
			Str("method", c.Request.Method).
			Str("path", path).
			Str("query", c.Request.URL.RawQuery).
			Int("status", status).
			Int("size", c.Writer.Size()).
			Dur("latency", time.Since(start)).
			Str("client_ip", c.ClientIP())

		if userAgent := c.Request.UserAgent(); userAgent != "" {
			event = event.Str("user_agent", userAgent)
		}
		event = withErrors(event, c.Errors)

		event.Msg("http request")
	}
}

// withErrors keeps the access log lean. A panic is only flagged, its details
// and its stack live in the error log written by the Recovery middleware.
func withErrors(event *zerolog.Event, errs []*gin.Error) *zerolog.Event {
	if event == nil {
		return event
	}

	for _, ginErr := range errs {
		if ginErr == nil {
			continue
		}

		var panicErr *PanicError
		if errors.As(ginErr.Err, &panicErr) {
			event = event.Bool("panic", true)
			continue
		}

		event = event.Str("error", ginErr.Err.Error())
	}

	return event
}

// stackFrame is one frame of a call stack.
type stackFrame struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

// stackFrames renders the current call stack as structured frames. It is
// used instead of debug.Stack, which produces a single blob of text that
// ends up as one unreadable escaped string in the log line.
func stackFrames(skip int) []stackFrame {
	pcs := make([]uintptr, 64)
	// 2 skips runtime.Callers and this helper, so the first frame is the
	// caller of stackFrames.
	count := runtime.Callers(skip+2, pcs)
	if count == 0 {
		return nil
	}

	frames := runtime.CallersFrames(pcs[:count])
	stack := make([]stackFrame, 0, count)
	for {
		frame, more := frames.Next()
		stack = append(stack, stackFrame{
			Function: frame.Function,
			File:     frame.File,
			Line:     frame.Line,
		})
		if !more {
			break
		}
	}

	return stack
}

// panicMessage renders the recovered value as a loggable message.
func panicMessage(recovered any) string {
	switch value := recovered.(type) {
	case nil:
		return ""
	case error:
		return value.Error()
	case string:
		return value
	default:
		return fmt.Sprintf("%v", value)
	}
}
