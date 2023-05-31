package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	autometrics "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"
	middleware "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// This should be `//go:generate autometrics` in practice. Those are hacks to get the example working, see
// README
//go:generate go run ../../../cmd/autometrics/main.go

var (
	Version = "development"
	Commit  = "n/a"
	Branch  string
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Everything in BuildInfo is optional.
	// You can also use any string variable whose value is
	// injected at build time by ldflags.
	autometrics.Init(
		nil,
		autometrics.DefBuckets,
		autometrics.BuildInfo{
			Version: Version,
			Commit:  Commit,
			Branch:  Branch,
		},
	)

	http.HandleFunc("/", errorable(indexHandler))
	// Wrapping a route in Autometrics middleware
	http.Handle("/random-error", middleware.Autometrics(
		http.HandlerFunc(randomErrorHandler),
		autometrics.WithSloName("API"),
		autometrics.WithAlertSuccess(90),
	))
	http.Handle("/metrics", promhttp.HandlerFor(
		prometheus.DefaultGatherer,
		promhttp.HandlerOpts{
			EnableOpenMetrics: true,
		}))

	log.Println("binding on http://localhost:62086")
	log.Fatal(http.ListenAndServe(":62086", nil))
}

// indexHandler handles the / route.
//
// It always succeeds and says hello.
//
//	autometrics:doc-start Generated documentation by Autometrics.
//
// # Autometrics
//
// # Prometheus
//
// View the live metrics for the `indexHandler` function:
//   - [Request Rate]
//   - [Error Ratio]
//   - [Latency (95th and 99th percentiles)]
//   - [Concurrent Calls]
//
// Or, dig into the metrics of *functions called by* `indexHandler`
//
//   - [Request Rate Callee]
//
//   - [Error Ratio Callee]
//
//     autometrics:doc-end Generated documentation by Autometrics.
//
// [Request Rate]: http://localhost:9090/graph?g0.expr=%23+Rate+of+calls+to+the+%60indexHandler%60+function+per+second%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%2C+version%2C+commit%29+%28rate%28function_calls_count%7Bfunction%3D%22indexHandler%22%7D%5B5m%5D%29+%2A+on+%28instance%2C+job%29+group_left%28version%2C+commit%29+last_over_time%28build_info%5B1s%5D%29%29&g0.tab=0
// [Error Ratio]: http://localhost:9090/graph?g0.expr=%23+Percentage+of+calls+to+the+%60indexHandler%60+function+that+return+errors%2C+averaged+over+5+minute+windows%0A%0A%28sum+by+%28function%2C+module%2C+version%2C+commit%29+%28rate%28function_calls_count%7Bfunction%3D%22indexHandler%22%2Cresult%3D%22error%22%7D%5B5m%5D%29+%2A+on+%28instance%2C+job%29+group_left%28version%2C+commit%29+last_over_time%28build_info%5B1s%5D%29%29%29+%2F+%28sum+by+%28function%2C+module%2C+version%2C+commit%29+%28rate%28function_calls_count%7Bfunction%3D%22indexHandler%22%7D%5B5m%5D%29+%2A+on+%28instance%2C+job%29+group_left%28version%2C+commit%29+last_over_time%28build_info%5B1s%5D%29%29%29&g0.tab=0
// [Latency (95th and 99th percentiles)]: http://localhost:9090/graph?g0.expr=%23+95th+and+99th+percentile+latencies+%28in+seconds%29+for+the+%60indexHandler%60+function%0A%0Alabel_replace%28histogram_quantile%280.99%2C+sum+by+%28le%2C+function%2C+module%2C+version%2C+commit%29+%28rate%28function_calls_duration_bucket%7Bfunction%3D%22indexHandler%22%7D%5B5m%5D%29+%2A+on+%28instance%2C+job%29+group_left%28version%2C+commit%29+last_over_time%28build_info%5B1s%5D%29%29%29%2C+%22percentile_latency%22%2C+%2299%22%2C+%22%22%2C+%22%22%29+or+label_replace%28histogram_quantile%280.95%2C+sum+by+%28le%2C+function%2C+module%2C+version%2C+commit%29+%28rate%28function_calls_duration_bucket%7Bfunction%3D%22indexHandler%22%7D%5B5m%5D%29+%2A+on+%28instance%2C+job%29+group_left%28version%2C+commit%29+last_over_time%28build_info%5B1s%5D%29%29%29%2C%22percentile_latency%22%2C+%2295%22%2C+%22%22%2C+%22%22%29&g0.tab=0
// [Concurrent Calls]: http://localhost:9090/graph?g0.expr=%23+Concurrent+calls+to+the+%60indexHandler%60+function%0A%0Asum+by+%28function%2C+module%2C+version%2C+commit%29+%28function_calls_concurrent%7Bfunction%3D%22indexHandler%22%7D+%2A+on+%28instance%2C+job%29+group_left%28version%2C+commit%29+last_over_time%28build_info%5B1s%5D%29%29&g0.tab=0
// [Request Rate Callee]: http://localhost:9090/graph?g0.expr=%23+Rate+of+function+calls+emanating+from+%60indexHandler%60+function+per+second%2C+averaged+over+5+minute+windows%0A%0Asum+by+%28function%2C+module%2C+version%2C+commit%29+%28rate%28function_calls_count%7Bcaller%3D%22main.indexHandler%22%7D%5B5m%5D%29+%2A+on+%28instance%2C+job%29+group_left%28version%2C+commit%29+last_over_time%28build_info%5B1s%5D%29%29&g0.tab=0
// [Error Ratio Callee]: http://localhost:9090/graph?g0.expr=%23+Percentage+of+function+emanating+from+%60indexHandler%60+function+that+return+errors%2C+averaged+over+5+minute+windows%0A%0A%28sum+by+%28function%2C+module%2C+version%2C+commit%29+%28rate%28function_calls_count%7Bcaller%3D%22main.indexHandler%22%2Cresult%3D%22error%22%7D%5B5m%5D%29+%2A+on+%28instance%2C+job%29+group_left%28version%2C+commit%29+last_over_time%28build_info%5B1s%5D%29%29%29+%2F+%28sum+by+%28function%2C+module%2C+version%2C+commit%29+%28rate%28function_calls_count%7Bcaller%3D%22main.indexHandler%22%7D%5B5m%5D%29+%2A+on+%28instance%2C+job%29+group_left%28version%2C+commit%29+last_over_time%28build_info%5B1s%5D%29%29%29&g0.tab=0
//
//autometrics:inst --slo "API" --latency-target 99 --latency-ms 100
func indexHandler(w http.ResponseWriter, r *http.Request) error {
	defer autometrics.Instrument(autometrics.PreInstrument(autometrics.NewContext(
		r.Context(),
		autometrics.WithConcurrentCalls(true),
		autometrics.WithCallerName(true),
		autometrics.WithSloName("API"),
		autometrics.WithAlertLatency(100000000*time.Nanosecond, 99),
	)), nil) //autometrics:defer

	time.Sleep(time.Duration(rand.Intn(200)) * time.Millisecond)

	fmt.Fprintf(w, "Hello, World!\n")

	return nil
}

var handlerError = errors.New("failed to handle request")

// randomErrorHandler handles the /random-error route.
//
// It returns an error around 90% of the time.
func randomErrorHandler(w http.ResponseWriter, r *http.Request) {
	isOk := rand.Intn(10) == 0

	if !isOk {
		http.Error(w, handlerError.Error(), http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	return
}

// errorable is a wrapper to allow using functions that return `error` in route handlers.
func errorable(handler func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := autometrics.WithNewTraceId(r.Context())
		if err := handler(w, r.WithContext(ctx)); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
