package prometheus // import "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"

import (
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	FunctionCallsCount      *prometheus.CounterVec
	FunctionCallsDuration   *prometheus.HistogramVec
	FunctionCallsConcurrent *prometheus.GaugeVec
	DefBuckets = autometrics.DefBuckets
)

const (
	FunctionCallsCountName      = "function_calls_count"
	FunctionCallsDurationName   = "function_calls_duration"
	FunctionCallsConcurrentName = "function_calls_concurrent"

	FunctionLabel          = "function"
	ModuleLabel            = "module"
	CallerLabel            = "caller"
	ResultLabel            = "result"
	TargetLatencyLabel     = "objective_latency_threshold"
	TargetSuccessRateLabel = "objective_percentile"
	SloNameLabel           = "objective_name"
)

// Init sets up the metrics required for autometrics' decorated functions and registers
// them to the argument registry.
//
// If the passed registry is nil, all the metrics are registered to the
// default global registry.
//
// Make sure that all the latency targets you want to use for SLOs are
// present in the histogramBuckets array, otherwise the alerts will fail
// to work (they will never trigger.)
func Init(reg *prometheus.Registry, histogramBuckets []float64) {
	FunctionCallsCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: FunctionCallsCountName,
	}, []string{FunctionLabel, ModuleLabel, CallerLabel, ResultLabel, TargetSuccessRateLabel, SloNameLabel})

	FunctionCallsDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    FunctionCallsDurationName,
		Buckets: histogramBuckets,
	}, []string{FunctionLabel, ModuleLabel, CallerLabel, TargetLatencyLabel, TargetSuccessRateLabel, SloNameLabel})

	FunctionCallsConcurrent = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: FunctionCallsConcurrentName,
	}, []string{FunctionLabel, ModuleLabel, CallerLabel})

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
