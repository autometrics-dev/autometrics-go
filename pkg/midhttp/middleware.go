// Package midhttp contains common types used in the downstream implementations of the middleware for net/http handlers.
package midhttp // import "github.com/autometrics-dev/autometrics-go/pkg/middleware/midhttp"

import (
	"net/http"
)

const RequestIdHeader = "X-Request-Id"

type autometricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// NewResponseWriter creates a new ResponseWriter that keeps track of the status
// code of the query for reporting purposes.
func NewResponseWriter(w http.ResponseWriter) *autometricsResponseWriter {
	return &autometricsResponseWriter{w, http.StatusOK}
}

func (amrw *autometricsResponseWriter) CurrentStatusCode() int {
	return amrw.statusCode
}

func (amrw *autometricsResponseWriter) WriteHeader(code int) {
	amrw.statusCode = code
	amrw.ResponseWriter.WriteHeader(code)
}
