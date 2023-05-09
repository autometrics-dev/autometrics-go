// Autometrics instruments annotated functions, and adds links in their doc comments to graphs of their live usage.
//
// By default, `autometrics` generates metric collection code for usage with the
// [Prometheus client library]. If you want to use [OpenTelemetry metrics]
// instead (with a prometheus exporter for the metrics), pass the `--otel` flag
// to the invocation.
//
// By default, when activating Service Level Objectives (SLOs) `autometrics`
// does not allow to use latency targets that are outside the default latencies
// defined in [autometrics.DefBuckets]. If you want to use custom latencies for
// your latency SLOs, pass the `--custom-latency` flag to the invocation.
//
// It is meant to be used in a Go generator context. As such, it takes mandatory arguments in the form of environment variables.
// You can also control the base URL of the prometheus instance in doc comments with an environment variable.
//         Note: If you do not use the custom latencies in the SLO, the allowed latencies (in seconds) are in [autometrics.DefBuckets].
//
// Check https://github.com/autometrics-dev/autometrics-go for more help (including examples) and information.
// Autometrics is built by Fiberplane -- https://autometrics.dev
//
// Usage: autometrics -f FILE_NAME -m MODULE_NAME [--prom_url PROMETHEUS_URL] [--otel] [--custom-latency]
//
// Options:
//   -f FILE_NAME           File to transform. [env: GOFILE]
//   -m MODULE_NAME         Module containing the file to transform. [env: GOPACKAGE]
//   --prom_url PROMETHEUS_URL
//                          Base URL of the Prometheus instance to generate links to. [default: http://localhost:9090, env: AM_PROMETHEUS_URL]
//   --otel                 Use [OpenTelemetry client library] to instrument code instead of default [Prometheus client library]. [default: false]
//   --custom-latency       Allow non-default latencies to be used in latency-based SLOs. [default: false]
//   --help, -h             display this help and exit
//   --version              display version and exit
//
// [Prometheus client library]: https://github.com/prometheus/client_golang
// [OpenTelemetry client library]: https://github.com/open-telemetry/opentelemetry-go
// [OpenTelemetry metrics]: https://opentelemetry.io/docs/instrumentation/go/
// [autometrics.DefBuckets]: https://godoc.org/github.com/autometrics-dev/autometrics-go/pkg/autometrics#DefBuckets
package main
