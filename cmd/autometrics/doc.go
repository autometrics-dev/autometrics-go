// Autometrics runs as Go generator and updates a source file to add usage queries and metric collection to annotated functions.
//
// As a Go generator, it relies on the environment variables `GOFILE` and
// `GOPACKAGE` to find the target file to edit.
//
// By default, `autometrics` generates metric collection code for usage with the
// [Prometheus client library]. If you want to use [OpenTelemetry metrics]
// instead (with a prometheus exporter for the metrics), pass the `-otel` flag
// to the invocation.
//
// By default, when activating Service Level Objectives (SLOs) `autometrics`
// does not allow to use latency targets that are outside the default latencies
// defined in [autometrics.DefBuckets]. If you want to use custom latencies for
// your latency SLOs, pass the `-custom-latency` flag to the invocation.
//
// By default, the generated links in the documentation point to a Prometheus
// instance at http://localhost:9090. You can use the environment variable
// `AM_PROMETHEUS_URL` to change the base URL in the documentation links.
//
// [Prometheus client library]: https://github.com/prometheus/client_golang
// [OpenTelemetry metrics]: https://opentelemetry.io/docs/instrumentation/go/
package main
