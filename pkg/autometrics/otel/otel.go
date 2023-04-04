package otel // import "github.com/autometrics-dev/autometrics-go/pkg/autometrics/otel"

import (
	"fmt"

	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/sdk/metric"
)

var (
	FunctionCallsCount      instrument.Int64Counter
	FunctionCallsDuration   instrument.Float64Histogram
	FunctionCallsConcurrent instrument.Int64UpDownCounter
)

const (
	FunctionCallsCountName      = "function.calls.count"
	FunctionCallsDurationName   = "function.calls.duration"
	FunctionCallsConcurrentName = "function.calls.concurrent"

	FunctionLabel          = "function"
	ModuleLabel            = "module"
	CallerLabel            = "caller"
	ResultLabel            = "result"
	TargetLatencyLabel     = "objective.latency_threshold"
	TargetSuccessRateLabel = "objective.percentile"
	SloNameLabel           = "objective.name"
)

// Init sets up the metrics required for autometrics' decorated functions and registers
// them to the Prometheus exporter
//
// Make sure that all the latency targets you want to use for SLOs are
// present in the histogramBuckets array, otherwise the alerts will fail
// to work (they will never trigger.)
func Init(meterName string, histogramBuckets []float64) error {
	exporter, err := prometheus.New()
	if err != nil {
		return fmt.Errorf("error initializing prometheus exporter: %w", err)
	}
	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	meter := provider.Meter(fmt.Sprintf("autometrics/%v", meterName))

	FunctionCallsCount, err = meter.Int64Counter(FunctionCallsCountName, instrument.WithDescription("The number of times the function has been called"))
	if err != nil {
		return fmt.Errorf("error initializing %v metric: %w", FunctionCallsCountName, err)
	}

	FunctionCallsDuration, err = meter.Float64Histogram(FunctionCallsDurationName, instrument.WithDescription("The duration of each function call"))
	if err != nil {
		return fmt.Errorf("error initializing %v metric: %w", FunctionCallsDurationName, err)
	}

	FunctionCallsConcurrent, err = meter.Int64UpDownCounter(FunctionCallsConcurrentName, instrument.WithDescription("The number of simultaneous calls of the function"))
	if err != nil {
		return fmt.Errorf("error initializing %v metric: %w", FunctionCallsConcurrentName, err)
	}

	return nil
}
