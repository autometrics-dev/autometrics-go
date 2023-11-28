package autometrics // import "github.com/autometrics-dev/autometrics-go/prometheus/autometrics"

import (
	"errors"

	am "github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics/log"
	"github.com/prometheus/client_golang/prometheus"
)

type initArguments struct {
	registry         *prometheus.Registry
	histogramBuckets []float64
	logger           log.Logger
	commit           string
	version          string
	branch           string
	service          string
	repoURL          string
	repoProvider     string
	pushCollectorURL string
	pushJobName      string
}

func defaultInitArguments() initArguments {
	return initArguments{
		histogramBuckets: am.DefBuckets,
		logger:           log.NoOpLogger{},
		pushJobName:      am.DefaultJobName(),
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

// WithRegistry sets the prometheus registry to use.
//
// The default is to use
// prometheus default registry.
func WithRegistry(registry *prometheus.Registry) InitOption {
	return initOptionFunc(func(initArgs *initArguments) error {
		initArgs.registry = registry
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
//
// Just as the prometheus library [push] configuration,
// "You can use just host:port or ip:port as url, in which case “http://” is
// added automatically. Alternatively, include the schema in the URL. However,
// do not include the “/metrics/jobs/…” part."
//
// The default value is an empty string, which also disables metric pushing.
//
// [push]: https://pkg.go.dev/github.com/prometheus/client_golang/prometheus/push#New
func WithPushCollectorURL(pushCollectorURL string) InitOption {
	return initOptionFunc(func(initArgs *initArguments) error {
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

// WithHistogramBuckets sets the buckets to use for the latency histograms.
//
// WARNING: your latency SLOs should always use thresolds that are _exactly_ a bucket boundary
// to ensure alert precision.
//
// The default value is [autometrics.DefBuckets]
func WithHistogramBuckets(histogramBuckets []float64) InitOption {
	return initOptionFunc(func(initArgs *initArguments) error {
		if len(histogramBuckets) == 0 {
			return errors.New("setting histogram buckets: the buckets for the histogram must have at least one value.")
		}
		initArgs.histogramBuckets = histogramBuckets
		return nil
	})
}
