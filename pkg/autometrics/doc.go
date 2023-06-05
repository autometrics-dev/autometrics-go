// Package autometrics provides automatic metric collection and reporting to functions.
//
// Depending on the implementation you want to use for metric collection (currently, [Prometheus] and [Open Telemetry] are supported), you can initialise the metrics collector, and then use a defer statement to automatically instrument a function body.
//
// The generator associated with autometrics generates the collection defer statement from argument in a directive comment, see
// the main project's [Readme] for more detail.
//
// [Readme]: https://github.com/autometrics-dev/autometrics-go
// [Prometheus]: https://godoc.org/github.com/autometrics-dev/autometrics-go/prometheus/autometrics
// [Open Telemetry]: https://godoc.org/github.com/autometrics-dev/autometrics-go/otel/autometrics
package autometrics
