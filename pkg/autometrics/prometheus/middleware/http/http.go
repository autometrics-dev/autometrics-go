package http // import "github.com/autometrics-dev/autometrics-go/pkg/prometheus/middleware/http"

import (
	"errors"
	"net/http"

	am "github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	prom "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"
	mid "github.com/autometrics-dev/autometrics-go/pkg/middleware/http"
)

func Autometrics(next http.HandlerFunc, opts ...am.Option) http.HandlerFunc {
	fn := func(rw http.ResponseWriter, r *http.Request) {
		arw := mid.NewResponseWriter(rw)
		ctx := prom.PreInstrument(prom.NewContext(r.Context(), opts...))

		// Compute then set the function name and module name labels
		ctx = am.SetCallInfo(ctx, am.ReflectFunctionModuleName(next))

		err := errors.New("Unfinished handler")

		defer prom.Instrument(ctx, &err)

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
