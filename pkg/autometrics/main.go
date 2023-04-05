package autometrics

import (
	"context"
	"fmt"
	"log"
	"time"
)

var (
	DefBuckets    = []float64{.005, .0075, .01, .025, .05, .075, .1, .25, .5, .75, 1, 2.5, 5, 7.5, 10}
	DefObjectives = []float64{90, 95, 99, 99.9}
)

// Implementation is an enumeration type for the
// possible implementations of metrics to use.
type Implementation int

const (
	PROMETHEUS Implementation = iota
	OTEL                      = iota
)

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
	// startTime is the start time of a single function execution.
	// Only autometrics.Instrument should read this value.
	// Only autometrics.PreInstrument should write this value.
	//
	// This value is only exported for the child packages "prometheus" and "otel"
	StartTime time.Time
	// callInfo contains all the relevant data for caller information.
	// Only autometrics.Instrument should read this value.
	// Only autometrics.PreInstrument should write/read this value.
	//
	// This value is only exported for the child packages "prometheus" and "otel"
	CallInfo CallInfo
	Context  context.Context
}

// CallInfo holds the information about the current function call and its parent names.
type CallInfo struct {
	// FuncName is name of the function being tracked.
	FuncName string
	// ModuleName is name of the module of the function being tracked.
	ModuleName string
	// ParentFuncName is name of the caller of the function being tracked.
	ParentFuncName string
	// ParentModuleName is name of the module of the caller of the function being tracked.
	ParentModuleName string
}

func NewContext() Context {
	return Context{
		TrackConcurrentCalls: true,
		TrackCallerName:      true,
		AlertConf:            nil,
		Context:              context.Background(),
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

		if c.AlertConf.Success != nil && c.AlertConf.Success.Objective <= 1 {
			log.Println("Warning: the target success rate is between 0 and 1, which is between 0 and 1%%. '1' is 1%% not 100%%!")
		}

		if c.AlertConf.Success != nil && c.AlertConf.Success.Objective > 100 {
			return fmt.Errorf("Cannot have a target success rate that is strictly greater than 100 (more than 100%%)")
		}

		if c.AlertConf.Success != nil && !contains(DefObjectives, c.AlertConf.Success.Objective) {
			return fmt.Errorf("Cannot have a target success rate that is not one of the predetermined ones by generated rules files (valid targets are %v)", DefObjectives)
		}

		if c.AlertConf.Latency != nil {
			if c.AlertConf.Latency.Objective <= 0 {
				return fmt.Errorf("Cannot have a target for latency SLO that is negative")
			}
			if c.AlertConf.Latency.Objective <= 1 {
				log.Println("Warning: the latency target success rate is between 0 and 1, which is between 0 and 1%%. '1' is 1%% not 100%%!")
			}
			if c.AlertConf.Latency.Objective > 100 {
				return fmt.Errorf("Cannot have a target for latency SLO that is greater than 100 (more than 100%%)")
			}
			if !contains(DefObjectives, c.AlertConf.Latency.Objective) {
				return fmt.Errorf("Cannot have a target for latency SLO that is not one of the predetermined in the generated rules files (valid targets are %v)", DefObjectives)
			}
			if c.AlertConf.Latency.Target <= 0 {
				return fmt.Errorf("Cannot have a target latency SLO threshold that is negative (responses expected before the query)")
			}
			if !contains(DefBuckets, c.AlertConf.Latency.Target.Seconds()) {
				return fmt.Errorf("Cannot have a target latency SLO threshold that does not match a bucket (valid threshold in seconds are %v)", DefBuckets)
			}
		}
	}

	return nil
}

func contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
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
