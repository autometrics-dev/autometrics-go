package autometrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var FunctionCallsCount *prometheus.Counter
var FunctionCallsDuration *prometheus.Histogram
var FunctionCallsConcurrent *prometheus.Gauge

// Init sets up the metrics required for autometrics' decorated functions
func Init() {
	functionCallsCounter := promauto.NewCounter(prometheus.CounterOpts{
		Name: "function_calls_count",
		ConstLabels: map[string]string{
			"function": "",
			"module":   "",
		},
	})

	functionCallDuration := promauto.NewHistogram(prometheus.HistogramOpts{
		Name: "function_calls_duration",
		ConstLabels: map[string]string{
			"function": "",
			"module":   "",
		},
	})

	functionCallsConcurrent := promauto.NewGauge(prometheus.GaugeOpts{
		Name: "function_calls_concurrent",
		ConstLabels: map[string]string{
			"function": "",
			"module":   "",
		},
	})

	// need to do it in two steps so the variable isn't homeless: https://stackoverflow.com/a/10536096/11494565
	FunctionCallsCount = &functionCallsCounter
	FunctionCallsDuration = &functionCallDuration
	FunctionCallsConcurrent = &functionCallsConcurrent
}
