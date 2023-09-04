package autometrics // import "github.com/autometrics-dev/autometrics-go/pkg/autometrics"

import (
	"time"
)

// Those are variables because we cannot have const of this type.
// These variables are not meant to be modified.
var (
	DefBuckets    = []float64{.005, .0075, .01, .025, .05, .075, .1, .25, .5, .75, 1, 2.5, 5, 7.5, 10}
	DefObjectives = []float64{90, 95, 99, 99.9}
)

const (
	AllowCustomLatenciesFlag = "-custom-latency"
)

// Implementation is an enumeration type for the
// possible implementations of metrics to use.
type Implementation int

const (
	PROMETHEUS Implementation = iota
	OTEL
)

const (
	// MiddlewareSpanIDKey is the key to use to index context in middlewares that do not use context.Context.
	MiddlewareSpanIDKey = "autometricsSpanID"
	// MiddlewareTraceIDKey is the key to use to index context in middlewares that do not use context.Context.
	MiddlewareTraceIDKey = "autometricsTraceID"
)

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

// BuildInfo holds the information about the current build of the instrumented code.
type BuildInfo struct {
	// Commit is the commit of the code.
	Commit string
	// Version is the version of the code.
	Version string
	// Branch is the branch of the build of the codebase.
	Branch string
	// ServiceName is the name of the service
	Service string
}

// PushConfiguration holds the information necessary to push metrics to an OTEL Collector.
type PushConfiguration struct {
	// URL of the collector to push to. It must be non-empty if this struct is built.
	// You can use just host:port or ip:port as url, in which case “http://” is added automatically.
	// Alternatively, include the schema in the URL. However, do not include the “/metrics/jobs/…” part.
	CollectorURL string

	// JobName is the name of the job to use when pushing metrics.
	//
	// Good values for this (taking into account replicated services) are for example:
	// - [GetOutboundIP] to automatically populate the jobname with the IP it's coming from,
	// - a [Uuid v1](https://pkg.go.dev/github.com/google/uuid#NewUUID)
	//   or a [Ulid](https://github.com/oklog/ulid) to get sortable IDs and keeping
	//   relationship to the machine generating metrics.
	// - a Uuid v4 if you don't want metrics leaking extra information about the IPs
	//
	// If JobName is empty here, autometrics will use the outbound IP if readable,
	// or a ulid here, see [DefaultJobName].
	JobName string
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
