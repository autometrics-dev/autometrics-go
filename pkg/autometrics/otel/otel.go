package otel // import "github.com/autometrics-dev/autometrics-go/pkg/autometrics/otel"

import (
	"fmt"

	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/aggregation"
)

var (
	FunctionCallsCount      instrument.Int64UpDownCounter
	FunctionCallsDuration   instrument.Float64Histogram
	FunctionCallsConcurrent instrument.Int64UpDownCounter
)

const (
	FunctionCallsCountName          = "function.calls.count"
	FunctionCallsCountPromName      = "function_calls_count"
	FunctionCallsDurationName       = "function.calls.duration"
	FunctionCallsDurationPromName   = "function_calls_duration"
	FunctionCallsConcurrentName     = "function.calls.concurrent"
	FunctionCallsConcurrentPromName = "function_calls_concurrent"

	FunctionLabel          = "function"
	ModuleLabel            = "module"
	CallerLabel            = "caller"
	ResultLabel            = "result"
	// Within the metric.Stream of a metric.View, it is only possible
	// to have AttributeFilter (attribute.Filter) that eventually just choose
	// to keep or remove a Key/Value pair from the source attribute set.
	// It is notably impossible to rename a key in an attribute. This is
	// why we will keep the 'objective_' prefix instead of using a more idiomatic
	// 'objective.' prefix, so that the exported metrics stay compatible with the
	// autometrics.rules.yml file.
	TargetLatencyLabel     = "objective_latency_threshold"
	TargetSuccessRateLabel = "objective_percentile"
	SloNameLabel           = "objective_name"
)

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
				Name: FunctionCallsDurationPromName,
				Aggregation: aggregation.ExplicitBucketHistogram{
					Boundaries: histogramBuckets,
				},
			},
		)),
		metric.WithView(metric.NewView(
			metric.Instrument{
				Name:  FunctionCallsCountName,
				Scope: instrumentation.Scope{Name: completeMeterName(meterName)},
			},
			metric.Stream{
				Name: FunctionCallsCountPromName,
			},
		)),
		metric.WithView(metric.NewView(
			metric.Instrument{
				Name:  FunctionCallsConcurrentName,
				Scope: instrumentation.Scope{Name: completeMeterName(meterName)},
			},
			metric.Stream{
				Name: FunctionCallsConcurrentPromName,
			},
		)),
	)
	meter := provider.Meter(completeMeterName(meterName))

	// We are using an UpDown counter instead of the natural Counter because with a monotonic counter
	// there is no way to remove the '_total' suffix from the exported metric name. This suffix
	// makes the exported metrics incompatible with the autometrics.rules.yml file.
	// Ref: https://github.com/open-telemetry/opentelemetry-go/blob/6b7e207953ce0a13d38da628a6aa48ad56058d2a/exporters/prometheus/exporter.go#L212-L215
	FunctionCallsCount, err = meter.Int64UpDownCounter(FunctionCallsCountName, instrument.WithDescription("The number of times the function has been called"))
	if err != nil {
		return fmt.Errorf("error initializing %v metric: %w", FunctionCallsCountName, err)
	}

	FunctionCallsDuration, err = meter.Float64Histogram(FunctionCallsDurationName, instrument.WithDescription("The duration of each function call, in seconds"))
	if err != nil {
		return fmt.Errorf("error initializing %v metric: %w", FunctionCallsDurationName, err)
	}

	FunctionCallsConcurrent, err = meter.Int64UpDownCounter(FunctionCallsConcurrentName, instrument.WithDescription("The number of simultaneous calls of the function"))
	if err != nil {
		return fmt.Errorf("error initializing %v metric: %w", FunctionCallsConcurrentName, err)
	}

	return nil
}
