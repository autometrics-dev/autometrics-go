package autometrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	FunctionCallsCount      *prometheus.CounterVec
	FunctionCallsDuration   *prometheus.HistogramVec
	FunctionCallsConcurrent *prometheus.GaugeVec
)

// Init sets up the metrics required for autometrics' decorated functions
func Init(reg *prometheus.Registry) {
	FunctionCallsCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "function_calls_count",
	}, []string{"function", "module", "result"})

	FunctionCallsDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name: "function_calls_duration",
	}, []string{"function", "module"})

	FunctionCallsConcurrent = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "function_calls_concurrent",
	}, []string{"function", "module"})

	reg.MustRegister(FunctionCallsCount)
	reg.MustRegister(FunctionCallsDuration)
	reg.MustRegister(FunctionCallsConcurrent)
}
