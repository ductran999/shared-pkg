// Package response defines the standard API envelope returned by
// services, plus constructors for the common cases.
//
// Example:
//
//	return response.OK(todos)
//	return response.Created(created)
//	return response.BadRequest("invalid id")
package response

// Envelope is the standard API response body.
type Envelope struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

// OK returns a 200 envelope with data.
func OK(data any, meta ...any) Envelope {
	return Envelope{Code: 200, Message: "OK", Data: data, Meta: mergeMeta(meta)}
}

// Created returns a 201 envelope with the created resource.
func Created(data any, meta ...any) Envelope {
	return Envelope{Code: 201, Message: "Created", Data: data, Meta: mergeMeta(meta)}
}

// BadRequest returns a 400 envelope. The message defaults to "Bad Request".
func BadRequest(msg ...string) Envelope {
	return Envelope{Code: 400, Message: firstOr(msg, "Bad Request")}
}

// Unauthorized returns a 401 envelope. The message defaults to "Unauthorized".
func Unauthorized(msg ...string) Envelope {
	return Envelope{Code: 401, Message: firstOr(msg, "Unauthorized")}
}

// Forbidden returns a 403 envelope. The message defaults to "Forbidden".
func Forbidden(msg ...string) Envelope {
	return Envelope{Code: 403, Message: firstOr(msg, "Forbidden")}
}

// NotFound returns a 404 envelope. The message defaults to "Not Found".
func NotFound(msg ...string) Envelope {
	return Envelope{Code: 404, Message: firstOr(msg, "Not Found")}
}

// InternalError returns a 500 envelope. The message defaults to "Internal Server Error".
func InternalError(msg ...string) Envelope {
	return Envelope{Code: 500, Message: firstOr(msg, "Internal Server Error")}
}

// firstOr returns the first non-empty message, or fallback.
func firstOr(msg []string, fallback string) string {
	for _, m := range msg {
		if m != "" {
			return m
		}
	}
	return fallback
}

// mergeMeta returns nil when no meta is given, the value as-is when a
// single one is given, and a merged map when several are given.
func mergeMeta(meta []any) any {
	switch len(meta) {
	case 0:
		return nil
	case 1:
		return meta[0]
	}
	out := make(map[string]any)
	for _, m := range meta {
		if mm, ok := m.(map[string]any); ok {
			for k, v := range mm {
				out[k] = v
			}
		}
	}
	return out
}
