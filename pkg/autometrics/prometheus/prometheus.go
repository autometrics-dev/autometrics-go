package prometheus // import "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"

import (
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	functionCallsCount      *prometheus.CounterVec
	functionCallsDuration   *prometheus.HistogramVec
	functionCallsConcurrent *prometheus.GaugeVec
	DefBuckets              = autometrics.DefBuckets
)

const (
	// FunctionCallsCountName is the name of the prometheus metric for the counter of calls to specific functions.
	FunctionCallsCountName = "function_calls_count"
	// FunctionCallsDurationName is the name of the prometheus metric for the duration histogram of calls to specific functions.
	FunctionCallsDurationName = "function_calls_duration"
	// FunctionCallsConcurrentName is the name of the prometheus metric for the number of simulateneously active calls to specific functions.
	FunctionCallsConcurrentName = "function_calls_concurrent"

	// FunctionLabel is the prometheus label that describes the function name.
	//
	// It is guaranteed that a (FunctionLabel, ModuleLabel) value pair is unique
	// and matches at most one function in the source code
	FunctionLabel = "function"
	// ModuleLabel is the prometheus label that describes the module name that contains the function.
	//
	// It is guaranteed that a (FunctionLabel, ModuleLabel) value pair is unique
	// and matches at most one function in the source code
	ModuleLabel = "module"
	// CallerLabel is the prometheus label that describes the name of the function that called
	// the current function.
	CallerLabel = "caller"
	// ResultLabel is the prometheus label that describes whether a function call is successful.
	ResultLabel = "result"
	// TargetLatencyLabel is the prometheus label that describes the latency to respect to match
	// the Service Level Objective.
	TargetLatencyLabel = "objective_latency_threshold"
	// TargetSuccessRateLabel is the prometheus label that describes the percentage of calls that
	// must succeed to match the Service Level Objective.
	//
	// In the case of latency objectives, it describes the percentage of
	// calls that must last less than the value in [TargetLatencyLabel].
	//
	// In the case of success objectives, it describes the percentage of calls
	// that must be successful (i.e. have their [ResultLabel] be 'ok').
	TargetSuccessRateLabel = "objective_percentile"
	// SloLabelName is the prometheus label that describes the name of the Service Level Objective.
	SloNameLabel = "objective_name"
)

// Instrumentor is an empty struct that implements [autometrics.Instrumentor] interface.
//
// TODO: Use this instrumentor in the API.
type Instrumentor struct{}

// var _ autometrics.Instrumentation = Instrumentor{}

// Init sets up the metrics required for autometrics' decorated functions and registers
// them to the argument registry.
//
// If the passed registry is nil, all the metrics are registered to the
// default global registry.
//
// Make sure that all the latency targets you want to use for SLOs are
// present in the histogramBuckets array, otherwise the alerts will fail
// to work (they will never trigger.)
func Init(reg *prometheus.Registry, histogramBuckets []float64) error {
	functionCallsCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: FunctionCallsCountName,
	}, []string{FunctionLabel, ModuleLabel, CallerLabel, ResultLabel, TargetSuccessRateLabel, SloNameLabel})

	functionCallsDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    FunctionCallsDurationName,
		Buckets: histogramBuckets,
	}, []string{FunctionLabel, ModuleLabel, CallerLabel, TargetLatencyLabel, TargetSuccessRateLabel, SloNameLabel})

	functionCallsConcurrent = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: FunctionCallsConcurrentName,
	}, []string{FunctionLabel, ModuleLabel, CallerLabel})

	if reg != nil {
		reg.MustRegister(functionCallsCount)
		reg.MustRegister(functionCallsDuration)
		reg.MustRegister(functionCallsConcurrent)
	} else {
		prometheus.DefaultRegisterer.MustRegister(functionCallsCount)
		prometheus.DefaultRegisterer.MustRegister(functionCallsDuration)
		prometheus.DefaultRegisterer.MustRegister(functionCallsConcurrent)
	}

	return nil
}
