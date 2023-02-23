package autometrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	FunctionCallsCount      *prometheus.CounterVec
	FunctionCallsDuration   *prometheus.HistogramVec
	FunctionCallsConcurrent *prometheus.GaugeVec
)

// Init sets up the metrics required for autometrics' decorated functions and registers
// them to the argument registry.
//
// If the passed registry is nil, all the metrics are registered to the
// default global registry.
func Init(reg *prometheus.Registry) {
	FunctionCallsCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "function_calls_count",
	}, []string{"function", "module", "caller", "result"})

	FunctionCallsDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "function_calls_duration",
	}, []string{"function", "module", "caller"})

	FunctionCallsConcurrent = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "function_calls_concurrent",
	}, []string{"function", "module", "caller"})

	if reg != nil {
		reg.MustRegister(FunctionCallsCount)
		reg.MustRegister(FunctionCallsDuration)
		reg.MustRegister(FunctionCallsConcurrent)
	} else {
		prometheus.DefaultRegisterer.MustRegister(FunctionCallsCount)
		prometheus.DefaultRegisterer.MustRegister(FunctionCallsDuration)
		prometheus.DefaultRegisterer.MustRegister(FunctionCallsConcurrent)
	}
}
