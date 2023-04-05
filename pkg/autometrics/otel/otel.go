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
	FunctionCallsCount      instrument.Int64Counter
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
	TargetLatencyLabel     = "objective.latency_threshold"
	TargetSuccessRateLabel = "objective.percentile"
	SloNameLabel           = "objective.name"
)

// // The default aggregation selector, but with the histogramBuckets instead
// // of the default ones.
// func makeAmAggregationSelector(histogramBuckets []float64) metric.AggregationSelector {
// 	return func(ik metric.InstrumentKind) aggregation.Aggregation {
// 		switch ik {
// 		case metric.InstrumentKindCounter, metric.InstrumentKindUpDownCounter, metric.InstrumentKindObservableCounter, metric.InstrumentKindObservableUpDownCounter:
// 			return aggregation.Sum{}
// 		case metric.InstrumentKindObservableGauge:
// 			return aggregation.LastValue{}
// 		case metric.InstrumentKindHistogram:
// 			return aggregation.ExplicitBucketHistogram{
// 				Boundaries: histogramBuckets,
// 				NoMinMax:   false,
// 			}
// 		}
// 		panic("unknown instrument kind")
// 	}
// }

func completeMeterName(meterName string) string {
	return fmt.Sprintf("autometric/%v", meterName)
}

// Init sets up the metrics required for autometrics' decorated functions and registers
// them to the Prometheus exporter
//
// Make sure that all the latency targets you want to use for SLOs are
// present in the histogramBuckets array, otherwise the alerts will fail
// to work (they will never trigger.)
func Init(meterName string, histogramBuckets []float64) error {
	exporter, err := prometheus.New(
		// prometheus.WithAggregationSelector(makeAmAggregationSelector(histogramBuckets)),
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

	FunctionCallsCount, err = meter.Int64Counter(FunctionCallsCountName, instrument.WithDescription("The number of times the function has been called"))
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
