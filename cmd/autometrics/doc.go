// Autometrics runs as Go generator and updates a source file to add usage queries and metric collection to annotated functions.
//
// As a Go generator, it relies on the environment variables `GOFILE` and `GOPACKAGE` to find the target file to edit.
//
// By default, `autometrics` generates metric collection code for usage with the [Prometheus client library]. If you want
// to use [OpenTelemetry metrics] instead (with a prometheus exporter for the metrics), pass the `-otel` flag to the
// invocation.
//
// [Prometheus client library]: https://github.com/prometheus/client_golang
// [OpenTelemetry metrics]: https://opentelemetry.io/docs/instrumentation/go/
package main
