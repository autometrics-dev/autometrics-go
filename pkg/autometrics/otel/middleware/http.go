package http // import "github.com/autometrics-dev/autometrics-go/pkg/autometrics/otel/middleware"

import (
	"errors"
	"net/http"

	am "github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	otel "github.com/autometrics-dev/autometrics-go/pkg/autometrics/otel"
	mid "github.com/autometrics-dev/autometrics-go/pkg/middleware/http"
)

func Autometrics(next http.Handler, opts ...am.Option) http.Handler {
	fn := func(rw http.ResponseWriter, r *http.Request) {
		arw := mid.NewResponseWriter(rw)
		ctx := otel.PreInstrument(otel.NewContext(r.Context(), opts...))
		// The Function name and modules are hardcoded to represent the HTTP route, instead
		// of autometrics.
		// The information about the handler's function name is not easily accessible (and what happens if
		// the handler is an anonymous function?)
		ctx = am.SetCallInfo(ctx, am.CallInfo{
			FuncName:   r.RequestURI,
			ModuleName: r.Host,
		})
		err := errors.New("Unfinished handler")

		defer otel.Instrument(ctx, &err)

		r = r.WithContext(ctx)
		next.ServeHTTP(arw, r)

		// Check the status code of the handler to reset the error before the Instrument deferred call
		ranges := am.GetValidHttpCodeRanges(ctx)
		for _, codeRange := range ranges {
			if codeRange.Contains(arw.CurrentStatusCode()) {
				err = nil
				break
			}
		}
	}

	return http.HandlerFunc(fn)
}
