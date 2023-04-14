// Package autometrics provides automatic metric collection and reporting to functions.
//
// Depending on the implementation you want to use for metric collection (currently, [autometrics/prometheus] and [autometrics/otel] are supported), you can initialise the metrics collector, and then use a defer statement to automatically instrument a function body.
//
// The generator associated with autometrics generates the collection defer statement from argument in a directive comment, see
// the main project's [Readme] for more detail.
//
// [Readme]: https://github.com/autometrics-dev/autometrics-go
package autometrics
