# OpenTelemetry example

This is simply an OpenTelemetry version of the [web](../web) example so most of
the README there also applies here.

The only difference is that the metrics implementation used is OpenTelemetry
with a Prometheus exporter instead of using a Prometheus only client crate.

You can notice the 3 differences that are mentionned in the top-level README:
- The amImpl import has been changed to `otel`
- The autometrics call in the Go generator has the `--otel` flag
- The `amImpl.Init` call uses a different first argument, with the name of the
  OpenTelemetry scope to use

## Quickstart

You can build and run the example by using the
`docker-compose.open-telemetry-example.yaml` file at the root of the repo.
