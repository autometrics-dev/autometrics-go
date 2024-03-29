package autometrics // import "github.com/autometrics-dev/autometrics-go/otel/autometrics"

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics/log"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/prometheus"
	instruments "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

var (
	functionCallsCount      instruments.Int64Counter
	functionCallsDuration   instruments.Float64Histogram
	functionCallsConcurrent instruments.Int64UpDownCounter
	buildInfo               instruments.Int64UpDownCounter
	DefBuckets              = autometrics.DefBuckets

	amCtx              context.Context
	exporterLock       sync.Mutex
	pushPeriodicReader *metric.PeriodicReader
)

const (
	// AutometricsSpecVersion is the version of the specification the library follows
	// The specifications can be found in https://github.com/autometrics-dev/autometrics-shared/tree/main/specs
	AutometricsSpecVersion = "1.0.0"

	// FunctionCallsCountName is the name of the openTelemetry metric for the counter of calls to specific functions.
	FunctionCallsCountName = "function.calls"
	// FunctionCallsDurationName is the name of the openTelemetry metric for the duration histogram of calls to specific functions.
	FunctionCallsDurationName = "function.calls.duration"
	// FunctionCallsConcurrentName is the name of the openTelemetry metric for the number of simulateneously active calls to specific functions.
	FunctionCallsConcurrentName = "function.calls.concurrent"
	// BuildInfo is the name of the openTelemetry metric for the version of the monitored codebase.
	BuildInfoName = "build_info"

	// FunctionLabel is the openTelemetry attribute that describes the function name.
	//
	// It is guaranteed that a (FunctionLabel, ModuleLabel) value pair is unique
	// and matches at most one function in the source code.
	FunctionLabel = "function"
	// ModuleLabel is the openTelemetry attribute that describes the module name that contains the function.
	//
	// It is guaranteed that a (FunctionLabel, ModuleLabel) value pair is unique
	// and matches at most one function in the source code.
	ModuleLabel = "module"
	// CallerFunctionLabel is the openTelemetry attribute that describes the name of the function that called
	// the current function.
	CallerFunctionLabel = "caller.function"
	// CallerModuleLabel is the openTelemetry attribute that describes the module of the function that called
	// the current function.
	CallerModuleLabel = "caller.module"
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
	// SloNameLabel is the openTelemetry attribute that describes the name of the Service Level Objective.
	SloNameLabel = "objective.name"

	// CommitLabel is the openTelemetry attribute that describes the commit of the monitored codebase.
	CommitLabel = "commit"
	// VersionLabel is the openTelemetry attribute that describes the version of the monitored codebase.
	VersionLabel = "version"
	// BranchLabel is the openTelemetry attribute that describes the branch of the build of the monitored codebase.
	BranchLabel = "branch"

	// RepositoryURLLabel is the openTelemetry attribute that describes the URL at which the repository containing
	// the monitored service can be found
	RepositoryURLLabel = "repository.url"
	// RepositoryProviderLabel is the openTelemetry attribute that describes the service provider for the monitored
	// service repository url
	RepositoryProviderLabel = "repository.provider"

	// AutometricsVersionLabel is the openTelemetry attribute that describes the version of the Autometrics specification
	// the library follows
	AutometricsVersionLabel = "autometrics.version"

	// ServiceNameLabel is the openTelemetry attribute that describes the name of the Service being monitored.
	ServiceNameLabel = "service.name"

	// JobNameLabel is the openTelemetry attribute that describes the job producing the metrics. It is
	// used when pushing OTLP metrics.
	JobNameLabel = "job"

	defaultPushPeriod  = 10 * time.Second
	defaultPushTimeout = 5 * time.Second
)

func completeMeterName(meterName string) string {
	return fmt.Sprintf("autometrics/%v", meterName)
}

// Logger is an interface for logging autometrics-related events.
//
// This is a reexport to allow using only the current package at call site.
type Logger = log.Logger

// This is a reexport to allow using only the current package at call site.
type PrintLogger = log.PrintLogger

// This is a reexport to allow using only the current package at call site.
type NoOpLogger = log.NoOpLogger

// Init sets up the metrics required for autometrics' decorated functions and registers
// them to the Prometheus exporter.
//
// After initialization, use the returned [context.CancelCauseFunc] to flush the last
// results and turn off metric collection for the remainder of the program's lifetime.
// It is a good candidate to be deferred in the usual case.
//
// Make sure that all the latency targets you want to use for SLOs are
// present in the histogramBuckets array, otherwise the alerts will fail
// to work (they will never trigger).
func Init(initOpts ...InitOption) (context.CancelCauseFunc, error) {
	var err error
	newCtx, cancelFunc := context.WithCancelCause(context.Background())
	amCtx = newCtx

	initArgs := defaultInitArguments()
	for _, initOpt := range initOpts {
		if err := initOpt.Apply(&initArgs); err != nil {
			return nil, fmt.Errorf("initializing options: %w", err)
		}
	}

	err = initArgs.Validate()
	if err != nil {
		return nil, fmt.Errorf("init options validation: %w", err)
	}

	autometrics.SetCommit(initArgs.commit)
	autometrics.SetVersion(initArgs.version)
	autometrics.SetBranch(initArgs.branch)
	autometrics.SetLogger(initArgs.logger)

	var pushExporter metric.Exporter
	if initArgs.HasPushEnabled() {
		pushExporter, err = initPushExporter(initArgs)
		if err != nil {
			return nil, fmt.Errorf("impossible to initialize OTLP exporter: %w", err)
		}
	}

	if serviceName, ok := os.LookupEnv(autometrics.AutometricsServiceNameEnv); ok {
		autometrics.SetService(serviceName)
	} else if serviceName, ok := os.LookupEnv(autometrics.OTelServiceNameEnv); ok {
		autometrics.SetService(serviceName)
	} else {
		autometrics.SetService(initArgs.service)
	}

	if repoURL, ok := os.LookupEnv(autometrics.AutometricsRepoURLEnv); ok {
		autometrics.SetRepositoryURL(repoURL)
	} else {
		autometrics.SetRepositoryURL(initArgs.repoURL)
	}
	if repoProvider, ok := os.LookupEnv(autometrics.AutometricsRepoProviderEnv); ok {
		autometrics.SetRepositoryProvider(repoProvider)
	} else {
		autometrics.SetRepositoryProvider(initArgs.repoProvider)
	}

	provider, err := initProvider(pushExporter, initArgs)
	if err != nil {
		return nil, err
	}
	meter := provider.Meter(completeMeterName(initArgs.meterName))

	functionCallsCount, err = meter.Int64Counter(FunctionCallsCountName, instruments.WithDescription("The number of times the function has been called"))
	if err != nil {
		return nil, fmt.Errorf("error initializing %v metric: %w", FunctionCallsCountName, err)
	}

	functionCallsDuration, err = meter.Float64Histogram(FunctionCallsDurationName, instruments.WithDescription("The duration of each function call, in seconds"))
	if err != nil {
		return nil, fmt.Errorf("error initializing %v metric: %w", FunctionCallsDurationName, err)
	}

	functionCallsConcurrent, err = meter.Int64UpDownCounter(FunctionCallsConcurrentName, instruments.WithDescription("The number of simultaneous calls of the function"))
	if err != nil {
		return nil, fmt.Errorf("error initializing %v metric: %w", FunctionCallsConcurrentName, err)
	}

	buildInfo, err = meter.Int64UpDownCounter(BuildInfoName, instruments.WithDescription("The information of the current build."))
	if err != nil {
		return nil, fmt.Errorf("error initializing %v metric: %w", BuildInfoName, err)
	}

	buildInfo.Add(amCtx, 1,
		instruments.WithAttributes(
			[]attribute.KeyValue{
				attribute.Key(CommitLabel).String(autometrics.GetCommit()),
				attribute.Key(VersionLabel).String(autometrics.GetVersion()),
				attribute.Key(BranchLabel).String(autometrics.GetBranch()),
				attribute.Key(ServiceNameLabel).String(autometrics.GetService()),
				attribute.Key(RepositoryProviderLabel).String(autometrics.GetRepositoryProvider()),
				attribute.Key(RepositoryURLLabel).String(autometrics.GetRepositoryURL()),
				attribute.Key(JobNameLabel).String(autometrics.GetPushJobName()),
				attribute.Key(AutometricsVersionLabel).String(AutometricsSpecVersion),
			}...))

	return cancelFunc, nil
}

// ForceFlush forces a flush of the metrics, in the case autometrics is pushing metrics to an OTLP collector.
//
// This function is a no-op if no push configuration has been setup in [Init], but will return an error if
// autometrics is not active (because this function is called before [Init] or after its shutdown function
// has been called).
func ForceFlush() error {
	if amCtx.Err() != nil {
		return fmt.Errorf("autometrics is not currently active: %w", amCtx.Err())
	}

	if pushPeriodicReader != nil {
		ctx, cancel := context.WithCancel(amCtx)
		defer cancel()
		if exporterLock.TryLock() {
			defer exporterLock.Unlock()
			if err := pushPeriodicReader.ForceFlush(ctx); err != nil {
				return fmt.Errorf("autometrics: opentelemetry: periodicReader: issue while flushing: %w\n", err)
			}
		}
	}

	return nil
}

func initProvider(pushExporter metric.Exporter, initArgs initArguments) (*metric.MeterProvider, error) {
	instrumentView := metric.Instrument{
		Name:  FunctionCallsDurationName,
		Scope: instrumentation.Scope{Name: completeMeterName(initArgs.meterName)},
	}
	streamView := metric.Stream{
		Aggregation: metric.AggregationExplicitBucketHistogram{
			Boundaries: initArgs.histogramBuckets,
		},
	}

	src, err := resource.Merge(
		resource.Default(),
		resource.Environment(),
	)
	if err != nil {
		src = resource.Default()
	}
	autometricsSrc, err := resource.Merge(
		src,
		resource.NewWithAttributes(
			semconv.SchemaURL,
			[]attribute.KeyValue{
				attribute.Key(semconv.ServiceNameKey).
					String(autometrics.GetService()),
				attribute.Key(semconv.ServiceInstanceIDKey).
					String(autometrics.GetPushJobName()),
			}...),
	)
	if err != nil {
		autometricsSrc = src
	}

	if pushExporter == nil {
		exporter, err := prometheus.New()
		if err != nil {
			return nil, fmt.Errorf("error initializing prometheus exporter: %w", err)
		}

		streamView.AttributeFilter = attribute.NewDenyKeysFilter(attribute.Key(JobNameLabel))

		metricView := metric.NewView(
			instrumentView,
			streamView,
		)

		return metric.NewMeterProvider(
			metric.WithReader(exporter),
			metric.WithView(metricView),
			metric.WithResource(autometricsSrc),
		), nil
	} else {
		autometrics.GetLogger().Debug("opentelemetry: setting up OTLP push configuration, pushing %s to %s\n",
			autometrics.GetPushJobName(),
			autometrics.GetPushJobURL(),
		)
		metricView := metric.NewView(
			instrumentView,
			streamView,
		)

		timeout := defaultPushTimeout
		interval := defaultPushPeriod

		readInitArgs := false
		if pushPeriod, ok := os.LookupEnv(autometrics.OTelPushPeriodEnv); ok {
			pushPeriodMs, err := strconv.ParseInt(pushPeriod, 10, 32)
			if err != nil {
				autometrics.GetLogger().Warn("opentelemetry: the push period environment variable has non-integer value, ignoring: %s", err)
				readInitArgs = true
			} else {
				interval = time.Duration(pushPeriodMs) * time.Millisecond
			}
		}
		if readInitArgs && initArgs.pushPeriod > 0 {
			interval = initArgs.pushPeriod
		}

		readInitArgs = false
		if pushTimeout, ok := os.LookupEnv(autometrics.OTelPushTimeoutEnv); ok {
			pushTimeoutMs, err := strconv.ParseInt(pushTimeout, 10, 32)
			if err != nil {
				autometrics.GetLogger().Warn("opentelemetry: the push timeout environment variable has non-integer value, ignoring: %s", err)
				readInitArgs = true
			} else {
				timeout = time.Duration(pushTimeoutMs) * time.Millisecond
			}
		}
		if readInitArgs && initArgs.pushTimeout > 0 {
			timeout = initArgs.pushTimeout
		}

		pushPeriodicReader = metric.NewPeriodicReader(
			pushExporter,
			metric.WithInterval(interval),
			metric.WithTimeout(timeout),
		)

		return metric.NewMeterProvider(
			metric.WithReader(pushPeriodicReader),
			metric.WithView(metricView),
			metric.WithResource(autometricsSrc),
		), nil
	}
}

func initPushExporter(initArgs initArguments) (metric.Exporter, error) {
	autometrics.GetLogger().Debug("opentelemetry: Init: detected push configuration")
	if initArgs.pushCollectorURL == "" {
		return nil, errors.New("invalid Push Configuration: the CollectorURL must be set.")
	}
	autometrics.SetPushJobURL(initArgs.pushCollectorURL)

	autometrics.SetPushJobName(initArgs.pushJobName)

	if initArgs.pushUseHTTP {
		options := []otlpmetrichttp.Option{
			otlpmetrichttp.WithEndpoint(autometrics.GetPushJobURL()),
		}

		if initArgs.pushInsecure {
			options = append(options, otlpmetrichttp.WithInsecure())
		}

		if initArgs.pushHeaders != nil {
			options = append(options, otlpmetrichttp.WithHeaders(initArgs.pushHeaders))
		}

		return otlpmetrichttp.New(
			amCtx,
			options...,
		)

	}

	// If we are here, we are using a gRPC exporter

	options := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(autometrics.GetPushJobURL()),
	}

	if initArgs.pushInsecure {
		options = append(options, otlpmetricgrpc.WithInsecure())
	}

	if initArgs.pushHeaders != nil {
		options = append(options, otlpmetricgrpc.WithHeaders(initArgs.pushHeaders))
	}

	return otlpmetricgrpc.New(
		amCtx,
		options...,
	)
}
