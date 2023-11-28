# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Go module
versioning](https://go.dev/doc/modules/version-numbers).

## [Unreleased](https://github.com/autometrics-dev/autometrics-go/compare/v0.9.0...main)

### Added

### Changed

- [All] The `Init` API has changed, to use arguments of type `InitOption` instead of using
  separate types. This means all default arguments do not need to be mentioned in the
  call of `Init`, and for the rest `autometrics` provides `With...` functions that allow
  customization.

### Deprecated

### Removed

### Fixed

- Fix a bug where the repository provider label would overwrite the repository URL label
  instead of using its own label.

### Security

## [0.9.0](https://github.com/autometrics-dev/autometrics-go/releases/tag/v0.9.0) 2023-11-17

The main goal of this release is to reach compatibility with 1.0.0 version of Autometrics
specification.

### Added

- [All] `autometrics` now also optionnally adds the repository where the code comes from
  in the `build_info` metric. `repository.url` and `repository.provider` can be either
  set in the `BuildInformation` structure when calling `autometrics.Init`, or by setting
  environment variables `AUTOMETRICS_REPOSITORY_URL` and `AUTOMETRICS_REPOSITORY_PROVIDER`,
  respectively.
- [All] `autometrics` now adds the version of the specification it follows in the `build_info`
  metric. `autometrics.version` label will point the version of autometrics specification the
  associated metrics will follow.

### Changed

- [All] `autometrics` now inserts 2 statements in each function it instruments. The context
  created to track the function execution is now put in a variable, so that callees in the
  function body can optionnally use the context to help with better trace/span/call graph
  tracking. If `autometrics` detects a `context.Context` it can use, it will shadow the
  context with the autometrics augmented one to reduce the changes to make in the code.
  Currently, `autometrics` can detect arguments of specific types in the function signatures
  to read from and replace contexts:
  + `context.Context` (read and replace)
  + `http.Request` (read only)
  + `buffalo.Request` (read and replace)
  + `gin.Context` (read only; _very_ experimental)
- [Generator] The generator now tries to keep going with instrumentation even if some instrumentation
  fails. On error, still _No file will be modified_, and the generator will exit with an error code
  and output all the collected errors instead of only the first one.
- [All] `autometrics.Init` function takes an additional argument, a logging interface to decide whether
  and how users want autometrics to log events. Passing `nil` for the argument value will use a "No Op"
  logger that does not do anything.

### Fixed

- [Generator] Fix a few generator crashes when instrumented functions have specific argument types.

### Security

- Update all dependencies to reduce dependabot alerts in
  + `google.golang.org/grpc`
  + `golang.org/x/net`

## [0.8.2](https://github.com/autometrics-dev/autometrics-go/releases/tag/v0.8.2) 2023-10-20

### Added

- [All] `autometrics` the go-generator binary accepts an `--inst-all` flag, to process all
  functions in the file even if they do not have any annotation
- [All] `autometrics` the go-generator binary accepts a `--rm-all` flag (that overrides the `--inst-all` flag)
  to remove autometrics from all annotated functions. This is useful to offboard autometrics after trying it:
  ```bash
  AM_RM_ALL=true go generate ./...  # Will remove all godoc and instrumentation calls
  sed -i '/\/\/.*autometrics/d' **/*.go   # A similar sed command will remove all comments containing 'autometrics'
  # A go linter/formatter of your choice can then clean up all unused imports to remove the automatically added ones.
  ```

## [0.8.1](https://github.com/autometrics-dev/autometrics-go/releases/tag/v0.8.1) 2023-10-13

### Changed

- [All] autometrics now reports the fully qualified package name and cleans up the pointer
  specifiers from reported pointer receiver methods

### Fixed

- [All] Fixed an issue that could lead to panics when running the generator on already-annotated files (#70)
- [All] autometrics will now add a single line doc comment with the function name if the function
  had no comment besides the autometrics directive. This allows IDEs to properly render the generated
  documentation even when going through gopls

## [0.8.0](https://github.com/autometrics-dev/autometrics-go/releases/tag/v0.8.0) 2023-09-18

### Added

- [Prometheus collector] `Init` call now takes an optional `PushConfiguration`
  argument that allows to push metrics to PushGateway, _on top_ of exposing the
  metrics to an endpoint.
  Using the push gateway is very useful to automatically send metrics when a
  web service is autoscaled and has replicas being created or deleted during
  normal operations.
  + the `CollectorURL` in the `PushConfiguration` is mandatory and follows the
    API of `prometheus/push.New`
  + the `JobName` parameter is optional, and `autometrics` will do a best effort
    to fill the value.

- [OpenTelemetry collector] `Init` call now takes an optional `PushConfiguration`
  argument that allows to push metrics in OTLP format, _instead_ of exposing the
  metrics to a Promtheus-format endpoint.
  Using an OTLP collector is very useful to automatically send metrics when a
  web service is autoscaled and has replicas being created or deleted during
  normal operations.
  + the `CollectorURL` in the `PushConfiguration` is mandatory
  + the `JobName` parameter is optional, and `autometrics` will do a best effort
    to fill the value.

### Changed

- [All] `autometrics.Init` function (for both `prometheus` and `otel` variants)
  take an extra optional argument. See the `Added` section above for details.
- [All] `autometrics.Init` function (for both `prometheus` and `otel` variants)
  now return a [CancelCauseFunc](https://pkg.go.dev/context#CancelCauseFunc)
  to cleanly shutdown (early or as a normal shutdown process) all metric collection.

## [0.7.0](https://github.com/autometrics-dev/autometrics-go/releases/tag/v0.7.0) 2023-09-04

### Added

- All metrics now have a `service_name` label, which can either be compiled in `Init` call, or
  filled at runtime from environment variables (in order of precedence):
  + `AUTOMETRICS_SERVICE_NAME`
  + `OTEL_SERVICE_NAME`

### Changed

- Function calls metric has been renamed from `function_calls_count_total` to `function_calls_total`
- Function calls duration histogram has been renamed from `function_calls_duration`
  to `function_calls_duration_seconds`
- Function caller label has been split from `caller` to `caller_function` and `caller_label`

## [0.6.1](https://github.com/autometrics-dev/autometrics-go/releases/tag/v0.6.1) 2023-07-24

### Added

- The Go generator now removes all the old defer statements in function bodies before re-adding
  only the necessary ones. This means calling `go generate` on a file that has no annotation
  at all effectively cleans up the whole file from autometrics.

### Changed

- Instead of returning an error when the go generator does not find the autometrics import
  in a file, it will add the needed import itself in the file.

### Fixed

- Code generation now works when `autometrics` is imported with the `_` alias
- Fix regression for latency data collection that only registered 0 microsecond latencies

## [0.6.0](https://github.com/autometrics-dev/autometrics-go/releases/tag/v0.6.0) 2023-06-06

### Added

- Support for tracing-like exemplars in metrics. When using the Prometheus
  instrumentation implementation, `trace_id`, `span_id`, and `parent_id` (for
  the parent span) are added as exemplars to the metrics when they are observed.
  Note that the Prometheus server needs to be [configured
  specifically](https://prometheus.io/docs/prometheus/latest/feature_flags/#exemplars-storage)
  to read the exemplars.
- Added new options to context constructors to manipulate the tracing
  information.
- Add a middleware for `net/http` handlers

### Changed

- The runtime autometrics.Context structure now can be used anywhere a
  `context.Context` can, and will automatically embed a copy of the context
  present in the annotated function arguments, when relevant.
- The Context constructor changed signature to allow inclusion of a parent
  context.
- Refactor imports to become more idiomatic. The imports changed as follows
```patch
import (
-	autometrics "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"
+	"github.com/autometrics-dev/autometrics-go/prometheus/autometrics"
-	middleware "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus/middleware/http"
+	"github.com/autometrics-dev/autometrics-go/prometheus/midhttp"
)
```
  You can use a global search/replace to change the URLs

### Deprecated

### Removed

### Fixed

### Security

## [0.5.0](https://github.com/autometrics-dev/autometrics-go/releases/tag/v0.5.0) 2023-05-17

### Added

- Changelog to summarize changes in a single place
- Pull Request template for the repository
- `--no-doc` argument to the generator to prevent the generator from
  generating documentation links in the doc comments in the given file
- `--no-doc` argument to the `autometrics:inst` directive to prevent
  the generator from generating links on specific functions

### Changed

- The `//autometrics:doc` directive has been renamed `//autometrics:inst`
- The type of counter for `function.calls.count` metric has been changed to
  a monotonic Int64 counter

### Deprecated

- The `//autometrics:doc` directive has been renamed `//autometrics:inst`

### Removed

### Fixed

### Security

## [0.4.0](https://github.com/autometrics-dev/autometrics-go/releases/tag/v0.4.0) 2023-05-09

### Added

- Build information (branch, commit, version) can optionally be added to metrics. All queries
  have been updated to use the new information when available
- The generator has proper `--version` and `--help` subcommands
-

### Changed

- Long flags now all take 2 `-`
```patch
- //go:generate autometrics -otel -custom-latency
+ //go:generate autometrics --otel --custom-latency
```

- Initialization of autometrics now takes a `BuildInfo` argument meant to be filled with the
  relevant build information. It can be default initialized if we want to opt-out of build
  information
```patch
- autometrics.Init(nil, autometrics.DefBuckets)
+ autometrics.Init(nil, autometrics.DefBuckets, autometrics.BuildInfo{})
```

### Deprecated

### Removed

### Fixed

### Security

## [0.3.1](https://github.com/autometrics-dev/autometrics-go/releases/tag/v0.3.1) 2023-04-20

### Added

- Github workflow to provide the Go Generator on release pages for all main architectures.

### Changed

### Deprecated

### Removed

### Fixed

### Security

## [0.3.0](https://github.com/autometrics-dev/autometrics-go/releases/tag/v0.3.0) 2023-04-14

### Added

- OpenTelemetry client library can be used to collect
  metrics instead of only working with prometheus client. The only difference is
  the implementation of metric collection; the OpenTelemetry implementation
  still uses the Prometheus exporter to expose the collected data, so the same
  documentation links actually work with the otel implementation, as shown in
  the new example directory
- Input validation. To prevent users from making SLOs that would not trigger the
  bundled alerts, there is now a verification step in the generator, that will
  error if a `-latency-ms` value (in a `//autometrics:doc` directive) does not
  match one of the values in the `autometrics.DefBuckets` default list. This
  assumes the user used `DefBuckets` in the `amImpl.Init` call in their code.
  There are situation where the default buckets aren't what we want, so we can
  change those buckets, and the target latencies in `//autometrics:doc`
  directives. In that case, the validation would trigger a false positive and
  prevent code generation. The generator now takes a `-custom-latency` flag to
  bypass the latency threshold verification, in the case the `Init` call does
  not use the default bucket values anyway.

### Changed

- Imports changed to accomodate choosing between Prometheus and OpenTelemetry
```diff
- import "github.com/autometrics-dev/autometrics-go/pkg/autometrics"
+ import amImpl "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"

[…]

- autometrics.Init(nil, autometrics.DefBuckets)
+ amImpl.Init(nil, amImpl.DefBuckets)
```
The generator will automatically replace all the other previous calls to `autometrics`

### Deprecated

### Removed

### Fixed

### Security

## [0.2.0](https://github.com/autometrics-dev/autometrics-go/releases/tag/v0.2.0) 2023-04-06

### Added

- Licenses

### Changed

### Deprecated

### Removed

### Fixed

- Alert generation rules now correctly deal with low traffic services

### Security

## [0.1.0](https://github.com/autometrics-dev/autometrics-go/releases/tag/v0.1.0) 2023-03-16

### Added

- Go generator to parse and work on files
- Generation of links to prometheus graphs within functions' doc comments
- Automatics alert generation in Prometheus
- Demo project that shows the usage of autometrics

### Changed

### Deprecated

### Removed

### Fixed

### Security
