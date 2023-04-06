package otel // import "github.com/autometrics-dev/autometrics-go/pkg/autometrics/otel"

import (
	"fmt"

	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"

	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregation"
)

var (
	functionCallsCount      instrument.Int64UpDownCounter
	functionCallsDuration   instrument.Float64Histogram
	functionCallsConcurrent instrument.Int64UpDownCounter
	DefBuckets              = autometrics.DefBuckets
)

const (
	// FunctionCallsCountName is the name of the openTelemetry metric for the counter of calls to specific functions.
	FunctionCallsCountName = "function.calls.count"
	// FunctionCallsDurationName is the name of the openTelemetry metric for the duration histogram of calls to specific functions.
	FunctionCallsDurationName = "function.calls.duration"
	// FunctionCallsConcurrentName is the name of the openTelemetry metric for the number of simulateneously active calls to specific functions.
	FunctionCallsConcurrentName = "function.calls.concurrent"

	// FunctionLabel is the openTelemetry attribute that describes the function name.
	//
	// It is guaranteed that a (FunctionLabel, ModuleLabel) value pair is unique
	// and matches at most one function in the source code
	FunctionLabel = "function"
	// ModuleLabel is the openTelemetry attribute that describes the module name that contains the function.
	//
	// It is guaranteed that a (FunctionLabel, ModuleLabel) value pair is unique
	// and matches at most one function in the source code
	ModuleLabel = "module"
	// CallerLabel is the openTelemetry attribute that describes the name of the function that called
	// the current function.
	CallerLabel = "caller"
	// ResultLabel is the openTelemetry attribute that describes whether a function call is successful.
	ResultLabel = "result"
	// TargetLatencyLabel is the openTelemetry attribute that describes the latency to respect to match
	// the Service Level Objective.
	TargetLatencyLabel = "objective.latency_threshold"
	// TargetSuccessRateLabel is the openTelemetry attribute that describes the percentage of calls that
	// must succeed to match the Service Level Objective.
	//
	// In the case of latency objectives, it describes the percentage of
	// calls that must last less than the value in [TargetLatencyLabel].
	//
	// In the case of success objectives, it describes the percentage of calls
	// that must be successful (i.e. have their [ResultLabel] be 'ok').
	TargetSuccessRateLabel = "objective.percentile"
	// SloLabelName is the openTelemetry attribute that describes the name of the Service Level Objective.
	SloNameLabel = "objective.name"
)

// Instrumentor is an empty struct that implements [autometrics.Instrumentation] interface.
//
// TODO: Use this instrumentor in the API.
type Instrumentor struct{}

// Compile time check that Instrumentor does all the necessary work
// var _ autometrics.Instrumentation = Instrumentor{}

func completeMeterName(meterName string) string {
	return fmt.Sprintf("autometrics/%v", meterName)
}

// Init sets up the metrics required for autometrics' decorated functions and registers
// them to the Prometheus exporter
//
// Make sure that all the latency targets you want to use for SLOs are
// present in the histogramBuckets array, otherwise the alerts will fail
// to work (they will never trigger.)
func Init(meterName string, histogramBuckets []float64) error {
	exporter, err := prometheus.New(
		// The units are removed from the exporter so that the names of the
		// exported metrics after the View rename are consistent with the
		// autometrics.rules.yml file
		prometheus.WithoutUnits(),
	)
	if err != nil {
		return fmt.Errorf("error initializing prometheus exporter: %w", err)
	}
	provider := metric.NewMeterProvider(
		metric.WithReader(exporter),
		metric.WithView(metric.NewView(
			metric.Instrument{
				Name:  FunctionCallsDurationName,
				Scope: instrumentation.Scope{Name: completeMeterName(meterName)},
			},
			metric.Stream{
				Aggregation: aggregation.ExplicitBucketHistogram{
					Boundaries: histogramBuckets,
				},
			},
		)),
	)
	meter := provider.Meter(completeMeterName(meterName))

	// We are using an UpDown counter instead of the natural Counter because with a monotonic counter
	// there is no way to remove the '_total' suffix from the exported metric name. This suffix
	// makes the exported metrics incompatible with the autometrics.rules.yml file.
	// Ref: https://github.com/open-telemetry/opentelemetry-go/blob/6b7e207953ce0a13d38da628a6aa48ad56058d2a/exporters/prometheus/exporter.go#L212-L215
	functionCallsCount, err = meter.Int64UpDownCounter(FunctionCallsCountName, instrument.WithDescription("The number of times the function has been called"))
	if err != nil {
		return fmt.Errorf("error initializing %v metric: %w", FunctionCallsCountName, err)
	}

	functionCallsDuration, err = meter.Float64Histogram(FunctionCallsDurationName, instrument.WithDescription("The duration of each function call, in seconds"))
	if err != nil {
		return fmt.Errorf("error initializing %v metric: %w", FunctionCallsDurationName, err)
	}

	functionCallsConcurrent, err = meter.Int64UpDownCounter(FunctionCallsConcurrentName, instrument.WithDescription("The number of simultaneous calls of the function"))
	if err != nil {
		return fmt.Errorf("error initializing %v metric: %w", FunctionCallsConcurrentName, err)
	}

	return nil
}
