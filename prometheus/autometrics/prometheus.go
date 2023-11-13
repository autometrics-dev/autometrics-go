package autometrics // import "github.com/autometrics-dev/autometrics-go/prometheus/autometrics"

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/prometheus/common/expfmt"
)

var (
	functionCallsCount      *prometheus.CounterVec
	functionCallsDuration   *prometheus.HistogramVec
	functionCallsConcurrent *prometheus.GaugeVec
	buildInfo               *prometheus.GaugeVec
	DefBuckets              = autometrics.DefBuckets

	amCtx      context.Context
	pusher     *push.Pusher
	pusherLock sync.Mutex
)

const (
	// AutometricsSpecVersion is the version of the specification the library follows
	// The specifications can be found in https://github.com/autometrics-dev/autometrics-shared/tree/main/specs
	AutometricsSpecVersion = "1.0.0"

	// FunctionCallsCountName is the name of the prometheus metric for the counter of calls to specific functions.
	FunctionCallsCountName = "function_calls_total"
	// FunctionCallsDurationName is the name of the prometheus metric for the duration histogram of calls to specific functions.
	FunctionCallsDurationName = "function_calls_duration_seconds"
	// FunctionCallsConcurrentName is the name of the prometheus metric for the number of simulateneously active calls to specific functions.
	FunctionCallsConcurrentName = "function_calls_concurrent"
	// BuildInfo is the name of the prometheus metric for the version of the monitored codebase.
	BuildInfoName = "build_info"

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
	// CallerFunctionLabel is the prometheus label that describes the name of the function that called
	// the current function.
	CallerFunctionLabel = "caller_function"
	// CallerModuleLabel is the prometheus label that describes the module of the function that called
	// the current function.
	CallerModuleLabel = "caller_module"
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
	// SloNameLabel is the prometheus label that describes the name of the Service Level Objective.
	SloNameLabel = "objective_name"

	// CommitLabel is the prometheus label that describes the commit of the monitored codebase.
	CommitLabel = "commit"
	// VersionLabel is the prometheus label that describes the version of the monitored codebase.
	VersionLabel = "version"
	// BranchLabel is the prometheus label that describes the branch of the build of the monitored codebase.
	BranchLabel = "branch"

	// RepositoryURLLabel is the prometheus label that describes the URL at which the repository containing
	// the monitored service can be found
	RepositoryURLLabel = "repository_url"
	// RepositoryProviderLabel is the prometheus label that describes the service provider for the monitored
	// service repository url
	RepositoryProviderLabel = "repository_provider"

	// AutometricsVersionLabel is the prometheus label that describes the version of the Autometrics specification
	// the library follows
	AutometricsVersionLabel = "autometrics_version"

	// ServiceNameLabel is the prometheus label that describes the name of the service being monitored
	ServiceNameLabel = "service_name"

	traceIdExemplar      = "trace_id"
	spanIdExemplar       = "span_id"
	parentSpanIdExemplar = "parent_id"
)

// BuildInfo holds meta information about the build of the instrumented code.
//
// This is a reexport of the autometrics type to allow [Init] to work with only
// the current (prometheus) package imported at the call site.
type BuildInfo = autometrics.BuildInfo

// PushConfiguration holds meta information about the push-to-collector configuration of the instrumented code.
//
// This is a reexport of the autometrics type to allow [Init] to work with only
// the current (prometheus) package imported at the call site.
//
// For the CollectorURL part, just as the prometheus library [push] configuration,
// "You can use just host:port or ip:port as url, in which case “http://” is
// added automatically. Alternatively, include the schema in the URL. However,
// do not include the “/metrics/jobs/…” part."
//
// [push]: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/push#New
type PushConfiguration = autometrics.PushConfiguration

// Init sets up the metrics required for autometrics' decorated functions and registers
// them to the argument registry.
//
// If the passed registry is nil, all the metrics are registered to the
// default global registry.
//
// After initialization, use the returned [context.CancelCauseFunc] to flush the last
// results and turn off metric collection for the remainder of the program's lifetime.
// It is a good candidate to be deferred in the usual case.
//
// Make sure that all the latency targets you want to use for SLOs are
// present in the histogramBuckets array, otherwise the alerts will fail
// to work (they will never trigger.)
func Init(reg *prometheus.Registry, histogramBuckets []float64, buildInformation BuildInfo, pushConfiguration *PushConfiguration) (context.CancelCauseFunc, error) {
	newCtx, cancelFunc := context.WithCancelCause(context.Background())
	amCtx = newCtx

	autometrics.SetCommit(buildInformation.Commit)
	autometrics.SetVersion(buildInformation.Version)
	autometrics.SetBranch(buildInformation.Branch)

	pusher = nil
	if pushConfiguration != nil {
		log.Printf("autometrics: Init: detected push configuration to %s", pushConfiguration.CollectorURL)

		if pushConfiguration.CollectorURL == "" {
			return nil, errors.New("invalid PushConfiguration: the CollectorURL must be set.")
		}
		autometrics.SetPushJobURL(pushConfiguration.CollectorURL)

		if pushConfiguration.JobName == "" {
			autometrics.SetPushJobName(autometrics.DefaultJobName())
		} else {
			autometrics.SetPushJobName(pushConfiguration.JobName)
		}

		pusher = push.
			New(autometrics.GetPushJobURL(), autometrics.GetPushJobName()).
			Format(expfmt.FmtText)

	}

	if serviceName, ok := os.LookupEnv(autometrics.AutometricsServiceNameEnv); ok {
		autometrics.SetService(serviceName)
	} else if serviceName, ok := os.LookupEnv(autometrics.OTelServiceNameEnv); ok {
		autometrics.SetService(serviceName)
	} else if buildInformation.Service != "" {
		autometrics.SetService(buildInformation.Service)
	}

	if repoURL, ok := os.LookupEnv(autometrics.AutometricsRepoURLEnv); ok {
		autometrics.SetRepositoryURL(repoURL)
	} else if buildInformation.RepositoryURL != "" {
		autometrics.SetRepositoryURL(buildInformation.RepositoryURL)
	}
	if repoProvider, ok := os.LookupEnv(autometrics.AutometricsRepoProviderEnv); ok {
		autometrics.SetRepositoryURL(repoProvider)
	} else if buildInformation.RepositoryProvider != "" {
		autometrics.SetRepositoryURL(buildInformation.RepositoryProvider)
	}

	functionCallsCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: FunctionCallsCountName,
	}, []string{FunctionLabel, ModuleLabel, CallerFunctionLabel, CallerModuleLabel, ResultLabel, TargetSuccessRateLabel, SloNameLabel, CommitLabel, VersionLabel, BranchLabel, ServiceNameLabel})

	functionCallsDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    FunctionCallsDurationName,
		Buckets: histogramBuckets,
	}, []string{FunctionLabel, ModuleLabel, CallerFunctionLabel, CallerModuleLabel, TargetLatencyLabel, TargetSuccessRateLabel, SloNameLabel, CommitLabel, VersionLabel, BranchLabel, ServiceNameLabel})

	functionCallsConcurrent = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: FunctionCallsConcurrentName,
	}, []string{FunctionLabel, ModuleLabel, CallerFunctionLabel, CallerModuleLabel, CommitLabel, VersionLabel, BranchLabel, ServiceNameLabel})

	buildInfo = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: BuildInfoName,
	}, []string{CommitLabel, VersionLabel, BranchLabel, ServiceNameLabel, RepositoryURLLabel, RepositoryProviderLabel, AutometricsVersionLabel})

	if reg != nil {
		reg.MustRegister(functionCallsCount)
		reg.MustRegister(functionCallsDuration)
		reg.MustRegister(functionCallsConcurrent)
		reg.MustRegister(buildInfo)
	} else {
		prometheus.DefaultRegisterer.MustRegister(functionCallsCount)
		prometheus.DefaultRegisterer.MustRegister(functionCallsDuration)
		prometheus.DefaultRegisterer.MustRegister(functionCallsConcurrent)
		prometheus.DefaultRegisterer.MustRegister(buildInfo)
	}

	buildInfo.With(prometheus.Labels{
		CommitLabel:             buildInformation.Commit,
		VersionLabel:            buildInformation.Version,
		BranchLabel:             buildInformation.Branch,
		ServiceNameLabel:        autometrics.GetService(),
		RepositoryURLLabel:      autometrics.GetRepositoryURL(),
		RepositoryProviderLabel: autometrics.GetRepositoryProvider(),
		AutometricsVersionLabel: AutometricsSpecVersion,
	}).Set(1)

	if pusher != nil {
		pusherLock.Lock()
		defer pusherLock.Unlock()

		if err := pusher.
			Collector(buildInfo).
			AddContext(amCtx); err != nil {
			return nil, fmt.Errorf("pushing metrics to gateway for initialization: %w", err)
		}
	}

	return cancelFunc, nil
}

// ForceFlush forces a flush of the metrics, in the case autometrics is pushing metrics to a Prometheus Push Gateway.
//
// This function is a no-op if no push configuration has been setup in [Init], but will return an error if
// autometrics is not active (because this function is called before [Init] or after its shutdown function
// has been called).
func ForceFlush() error {
	if amCtx.Err() != nil {
		return fmt.Errorf("autometrics is not currently active: %w", amCtx.Err())
	}

	if pusher != nil {
		ctx, cancel := context.WithCancel(amCtx)
		defer cancel()
		if pusherLock.TryLock() {
			defer pusherLock.Unlock()
			localPusher := push.
				New(autometrics.GetPushJobURL(), autometrics.GetPushJobName()).
				Format(expfmt.FmtText).
				Collector(functionCallsCount).
				Collector(functionCallsDuration).
				Collector(functionCallsConcurrent)
			if err := localPusher.
				AddContext(ctx); err != nil {
				return fmt.Errorf("pushing metrics to gateway: %w", err)
			}
		}
	}

	return nil
}
