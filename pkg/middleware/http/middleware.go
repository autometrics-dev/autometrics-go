package middleware // import "github.com/autometrics-dev/autometrics-go/pkg/middleware/http/middleware"

import (
	"net/http"

	am "github.com/autometrics-dev/autometrics-go/pkg/autometrics"
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

func Autometrics(next http.Handler) http.Handler {
	fn := func(rw http.ResponseWriter, r *http.Request) {
		arw := NewResponseWriter(rw)
		ctx := am.NewContext(r.Context())
		r = r.WithContext(ctx)
		next.ServeHTTP(arw, r)
		// TODO: This is where we would check the range of responses that are considered
		// ok or not, on arw.statusCode
	}

	return http.HandlerFunc(fn)
}
