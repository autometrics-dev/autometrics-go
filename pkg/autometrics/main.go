package autometrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	FunctionCallsCount      *prometheus.CounterVec
	FunctionCallsDuration   *prometheus.HistogramVec
	FunctionCallsConcurrent *prometheus.GaugeVec
	DefBuckets              = []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10}
)

const (
	FunctionCallsCountName      = "function_calls_count"
	FunctionCallsDurationName   = "function_calls_duration"
	FunctionCallsConcurrentName = "function_calls_concurrent"

	FunctionLabel          = "function"
	ModuleLabel            = "module"
	CallerLabel            = "caller"
	ResultLabel            = "result"
	TargetLatencyLabel     = "target_latency"
	TargetSuccessRateLabel = "objective"
	SloNameLabel           = "slo_name"
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

// Context holds the configuration
// to instrument properly a function.
//
// This can be viewed as a context for the instrumentation calls
type Context struct {
	// TrackConcurrentCalls triggers the collection of the gauge for concurrent calls of the function.
	TrackConcurrentCalls bool
	// TrackCallerName adds a label with the caller name in all the collected metrics.
	TrackCallerName bool
	// AlertConf is an optional configuration to add alerting capabilities to the metrics.
	AlertConf *AlertConfiguration
}

func NewContext() Context {
	return Context{
		TrackConcurrentCalls: true,
		TrackCallerName:      true,
		AlertConf:            nil,
	}
}

func (c Context) Validate() error {
	if c.AlertConf != nil {
		if c.AlertConf.ServiceName == "" {
			return fmt.Errorf("Cannot have an AlertConfiguration without a service name")
		}

		if c.AlertConf.Success != nil && c.AlertConf.Success.Objective <= 0 {
			return fmt.Errorf("Cannot have a target success rate that is negative")
		}

		if c.AlertConf.Success != nil && c.AlertConf.Success.Objective > 1 {
			return fmt.Errorf("Cannot have a target success rate that is strictly greater than 1 (more than 100%%)")
		}

		if c.AlertConf.Latency != nil {
			if c.AlertConf.Latency.Objective <= 0 {
				return fmt.Errorf("Cannot have a target for latency SLO that is negative")
			}
			if c.AlertConf.Latency.Objective > 1 {
				return fmt.Errorf("Cannot have a target for latency SLO that is greater than 1 (more than 100%%)")
			}
			if c.AlertConf.Latency.Target < 0 {
				return fmt.Errorf("Cannot have a target latency SLO threshold that is negative (responses expected before the query)")
			}
		}
	}

	return nil
}

// AlertConfiguration is the configuration for autometric alerting.
type AlertConfiguration struct {
	// ServiceName is the name of the Service that will appear in the alerts.
	ServiceName string
	// Latency is an optional latency target for the function
	Latency *LatencySlo
	// Success is an optional success rate target for the function
	Success *SuccessSlo
}

// LatencySlo is an objective for latency
type LatencySlo struct {
	// Target is the maximum allowed latency for the endpoint.
	Target time.Duration
	// Objective is the success rate allowed for the given latency, from 0 to 1.
	Objective float64
}

// SuccessSlo is an objective for the success rate of the function
type SuccessSlo struct {
	// Objective is the success rate allowed for the given function, from 0 to 1.
	Objective float64
}
