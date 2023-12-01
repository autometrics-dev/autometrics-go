package autometrics // import "github.com/autometrics-dev/autometrics-go/prometheus/autometrics"

import (
	"errors"
	"strings"
	"time"

	am "github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics/log"
)

type initArguments struct {
	meterName        string
	histogramBuckets []float64
	logger           log.Logger
	commit           string
	version          string
	branch           string
	service          string
	repoURL          string
	repoProvider     string
	pushCollectorURL string
	pushPeriod       time.Duration
	pushTimeout      time.Duration
	pushUseHTTP      bool
	pushHeaders      map[string]string
	pushInsecure     bool
	pushJobName      string
}

func defaultInitArguments() initArguments {
	return initArguments{
		histogramBuckets: am.DefBuckets,
		logger:           log.NoOpLogger{},
		pushJobName:      am.DefaultJobName(),
		pushPeriod:       defaultPushPeriod,
		pushTimeout:      defaultPushTimeout,
		pushUseHTTP:      false,
		pushInsecure:     false,
	}
}

func (initArgs initArguments) Validate() error {
	return nil
}

func (initArgs initArguments) HasPushEnabled() bool {
	return initArgs.pushCollectorURL != ""
}

type InitOption interface {
	Apply(*initArguments) error
}

type initOptionFunc func(*initArguments) error

func (fn initOptionFunc) Apply(initArgs *initArguments) error {
	return fn(initArgs)
}

// WithMeterName sets the name of the meter to use for opentelemetry metrics.
// The name will be prefixed with "autometrics/" to help figure out the origin.
//
// The default value is an empty string
func WithMeterName(currentMeterName string) InitOption {
	return initOptionFunc(func(initArgs *initArguments) error {
		initArgs.meterName = currentMeterName
		return nil
	})
}

// WithLogger sets the logger to use when initializing autometrics.
//
// The default logger is a no-op logger that will never log
// autometrics-specific events.
func WithLogger(logger log.Logger) InitOption {
	return initOptionFunc(func(initArgs *initArguments) error {
		initArgs.logger = logger
		return nil
	})
}

// WithCommit sets the commit of the codebase to export with the metrics.
//
// The default value is an empty string.
func WithCommit(currentCommit string) InitOption {
	return initOptionFunc(func(initArgs *initArguments) error {
		initArgs.commit = currentCommit
		return nil
	})
}

// WithVersion sets the version of the codebase to export with the metrics.
//
// The default value is an empty string.
func WithVersion(currentVersion string) InitOption {
	return initOptionFunc(func(initArgs *initArguments) error {
		initArgs.version = currentVersion
		return nil
	})
}

// WithBranch sets the name of the branch to export with the metrics.
//
// The default value is an empty string.
func WithBranch(currentBranch string) InitOption {
	return initOptionFunc(func(initArgs *initArguments) error {
		initArgs.branch = currentBranch
		return nil
	})
}

// WithService sets the name of the current service, to export with the metrics.
//
// The default value is an empty string.
func WithService(currentService string) InitOption {
	return initOptionFunc(func(initArgs *initArguments) error {
		initArgs.service = currentService
		return nil
	})
}

// WithRepoURL sets the URL of the repository containing the codebase being instrumented.
//
// The default value is an empty string.
func WithRepoURL(currentRepoURL string) InitOption {
	return initOptionFunc(func(initArgs *initArguments) error {
		initArgs.repoURL = currentRepoURL
		return nil
	})
}

// WithRepoProvider sets the provider of the repository containing the codebase being instrumented.
//
// The default value is an empty string.
func WithRepoProvider(currentRepoProvider string) InitOption {
	return initOptionFunc(func(initArgs *initArguments) error {
		initArgs.repoProvider = currentRepoProvider
		return nil
	})
}

// WithPushCollectorURL enables Pushing metrics to a remote location, and sets the URL of the
// collector to target.
// You can use just host:port or ip:port as url, in which case “http://” is added automatically.
// Alternatively, include the schema in the URL. However, do not include the “/metrics/jobs/…” part.
//
// The default value is an empty string, which also disables metric pushing.
func WithPushCollectorURL(pushCollectorURL string) InitOption {
	return initOptionFunc(func(initArgs *initArguments) error {
		if strings.Contains(pushCollectorURL, "/metrics/jobs") {
			return errors.New("set push collector URL: the URL should not contain the /metrics/jobs part")
		}
		initArgs.pushCollectorURL = pushCollectorURL
		return nil
	})
}

// WithPushJobName sets the name of job to use when pushing metrics.
//
// Good values for this (taking into account replicated services) are for example:
//   - The (internal) IP the job is coming from,
//   - a [Uuid v1](https://pkg.go.dev/github.com/google/uuid#NewUUID)
//     or a [Ulid](https://github.com/oklog/ulid) to get sortable IDs and keeping
//     relationship to the machine generating metrics.
//   - a Uuid v4 if you don't want metrics leaking extra information about the IPs
//
// The default value is an empty string, which will make autometrics generate a Ulid
func WithPushJobName(pushJobName string) InitOption {
	return initOptionFunc(func(initArgs *initArguments) error {
		initArgs.pushJobName = pushJobName
		return nil
	})
}

// WithPushPeriod sets the duration between consecutive metrics pushes.
//
// The standard `OTEL_METRIC_EXPORT_INTERVAL` environment variable overrides
// this initialization argument.
//
// The default value is 10 seconds.
func WithPushPeriod(pushPeriod time.Duration) InitOption {
	return initOptionFunc(func(initArgs *initArguments) error {
		initArgs.pushPeriod = pushPeriod
		return nil
	})
}

// WithPushTimeout sets the timeout duration of a single metric push
//
// The standard `OTEL_METRIC_EXPORT_TIMEOUT` environment variable overrides
// this initialization argument.
//
// The default value is 5 seconds.
func WithPushTimeout(pushTimeout time.Duration) InitOption {
	return initOptionFunc(func(initArgs *initArguments) error {
		initArgs.pushTimeout = pushTimeout
		return nil
	})
}

// WlthPushHTTP sets the metrics pushing mechanism to use the HTTP format over gRPC
//
// The default value is to use gRPC.
func WithPushHTTP() InitOption {
	return initOptionFunc(func(initArgs *initArguments) error {
		initArgs.pushUseHTTP = true
		return nil
	})
}

// WlthPushInsecure allows to use insecure (clear text) connections between the
// codebase and the metrics collector.
//
// The default value is to use secure channels only.
func WithPushInsecure() InitOption {
	return initOptionFunc(func(initArgs *initArguments) error {
		initArgs.pushInsecure = true
		return nil
	})
}

// WithPushHeaders allows adding headers to the payload of metrics when pushed to the
// collector (for BasicAuth authentication for example)
//
// The default value is empty.
func WithPushHeaders(headers map[string]string) InitOption {
	return initOptionFunc(func(initArgs *initArguments) error {
		initArgs.pushHeaders = headers
		return nil
	})
}

// WithHistogramBuckets sets the buckets to use for the latency histograms.
//
// WARNING: your latency SLOs should always use thresolds that are _exactly_ a bucket boundary
// to ensure alert precision.
//
// The default value is [autometrics.DefBuckets]
func WithHistogramBuckets(histogramBuckets []float64) InitOption {
	return initOptionFunc(func(initArgs *initArguments) error {
		if len(histogramBuckets) == 0 {
			return errors.New("setting histogram buckets: the histogram buckets should have at least 1 value.")
		}
		initArgs.histogramBuckets = histogramBuckets
		return nil
	})
}
