package autometrics // import "github.com/autometrics-dev/autometrics-go/otel/autometrics"

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"

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

// BuildInfo holds meta information about the build of the instrumented code.
//
// This is a reexport of the autometrics type to allow [Init] to work with only
// the current (otel) package imported at the call site.
type BuildInfo = autometrics.BuildInfo

// PushConfiguration holds meta information about the push-to-collector configuration of the instrumented code.
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

	// UseHttp forces the use of HTTP over GRPC as a protocol to push metrics to a
	// collector.
	//
	// It defaults to false, meaning that by default, autometrics will push GRPC
	// metrics to the collector.
	UseHttp bool

	// Headers is a map of headers to add to the payload when pushing metrics.
	Headers map[string]string

	// IsInsecure disables client transport security (such as TLS).
	IsInsecure bool

	// Period is the interval at which the metrics will be pushed to the collector.
	//
	// This option overrides any value set for the OTEL_METRIC_EXPORT_INTERVAL environment variable.
	//
	// It should be greater than Timeout, and if the Period is non-positive, a default
	// value of 10 seconds will be used.
	Period time.Duration

	// Timeout is the timeout duration for metrics pushes to the collector.
	//
	// This option overrides any value set for the OTEL_METRIC_EXPORT_TIMEOUT environment variable.
	//
	// It should be smaller than Period, and if the Timeout is non-positive, a default
	// value of 5 seconds will be used.
	Timeout time.Duration
}

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
func Init(meterName string, histogramBuckets []float64, buildInformation BuildInfo, pushConfiguration *PushConfiguration) (context.CancelCauseFunc, error) {
	var err error
	newCtx, cancelFunc := context.WithCancelCause(context.Background())
	amCtx = newCtx

	autometrics.SetCommit(buildInformation.Commit)
	autometrics.SetVersion(buildInformation.Version)
	autometrics.SetBranch(buildInformation.Branch)

	var pushExporter metric.Exporter
	if pushConfiguration != nil {
		pushExporter, err = initPushExporter(pushConfiguration)
		if err != nil {
			return nil, fmt.Errorf("impossible to initialize OTLP exporter: %w", err)
		}
	}

	if serviceName, ok := os.LookupEnv(autometrics.AutometricsServiceNameEnv); ok {
		autometrics.SetService(serviceName)
	} else if serviceName, ok := os.LookupEnv(autometrics.OTelServiceNameEnv); ok {
		autometrics.SetService(serviceName)
	} else if buildInformation.Service != "" {
		autometrics.SetService(buildInformation.Service)
	}

	provider, err := initProvider(pushExporter, pushConfiguration, meterName, histogramBuckets)
	if err != nil {
		return nil, err
	}
	meter := provider.Meter(completeMeterName(meterName))

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
				attribute.Key(CommitLabel).String(buildInformation.Commit),
				attribute.Key(VersionLabel).String(buildInformation.Version),
				attribute.Key(BranchLabel).String(buildInformation.Branch),
				attribute.Key(ServiceNameLabel).String(autometrics.GetService()),
				attribute.Key(JobNameLabel).String(autometrics.GetPushJobName()),
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

func initProvider(pushExporter metric.Exporter, pushConfiguration *PushConfiguration, meterName string, histogramBuckets []float64) (*metric.MeterProvider, error) {
	instrumentView := metric.Instrument{
		Name:  FunctionCallsDurationName,
		Scope: instrumentation.Scope{Name: completeMeterName(meterName)},
	}
	streamView := metric.Stream{
		Aggregation: metric.AggregationExplicitBucketHistogram{
			Boundaries: histogramBuckets,
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
		log.Printf("autometrics: opentelemetry: setting up OTLP push configuration, pushing %s to %s\n",
			autometrics.GetPushJobName(),
			autometrics.GetPushJobURL(),
		)
		metricView := metric.NewView(
			instrumentView,
			streamView,
		)

		timeout := defaultPushTimeout
		interval := defaultPushPeriod

		if pushConfiguration.Period > 0 {
			interval = pushConfiguration.Period
		}
		if pushConfiguration.Timeout > 0 {
			timeout = pushConfiguration.Timeout
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

func initPushExporter(pushConfiguration *PushConfiguration) (metric.Exporter, error) {
	log.Println("autometrics: opentelemetry: Init: detected push configuration")
	if pushConfiguration.CollectorURL == "" {
		return nil, errors.New("invalid PushConfiguration: the CollectorURL must be set.")
	}
	autometrics.SetPushJobURL(pushConfiguration.CollectorURL)

	if pushConfiguration.JobName == "" {
		autometrics.SetPushJobName(autometrics.DefaultJobName())
	} else {
		autometrics.SetPushJobName(pushConfiguration.JobName)
	}

	if pushConfiguration.UseHttp {
		options := []otlpmetrichttp.Option{
			otlpmetrichttp.WithEndpoint(autometrics.GetPushJobURL()),
		}

		if pushConfiguration.IsInsecure {
			options = append(options, otlpmetrichttp.WithInsecure())
		}

		if pushConfiguration.Headers != nil {
			options = append(options, otlpmetrichttp.WithHeaders(pushConfiguration.Headers))
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

	if pushConfiguration.IsInsecure {
		options = append(options, otlpmetricgrpc.WithInsecure())
	}

	if pushConfiguration.Headers != nil {
		options = append(options, otlpmetricgrpc.WithHeaders(pushConfiguration.Headers))
	}

	return otlpmetricgrpc.New(
		amCtx,
		options...,
	)
}
