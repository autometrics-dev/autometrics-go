package autometrics // import "github.com/autometrics-dev/autometrics-go/pkg/autometrics"

import (
	"fmt"
	"net/http"
)

type autometricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *autometricsResponseWriter {
	return &autometricsResponseWriter{w, http.StatusOK}
}

func (amrw *autometricsResponseWriter) WriteHeader(code int) {
	amrw.statusCode = code
	amrw.ResponseWriter.WriteHeader(code)
}

// HasHttpError returns non-nil if the response writer is an autometrics wrapper,
// and if the inner status code is outside of the 200-399 range.
func HasHttpError(rw http.ResponseWriter) error {
	if amrw, ok := rw.(*autometricsResponseWriter); ok {
		if amrw.statusCode < 200 || amrw.statusCode >= 400 {
			return fmt.Errorf("HTTP error %s (%d)", http.StatusText(amrw.statusCode), amrw.statusCode)
		}
	}
	return nil
}
