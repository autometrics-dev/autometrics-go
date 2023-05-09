# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Go module
versioning](https://go.dev/doc/modules/version-numbers).

## [Unreleased](https://github.com/autometrics-dev/autometrics-go/compare/v0.4.0...main)

### Added

- Changelog to summarize changes in a single place
- Pull Request template for the repository

### Changed

### Deprecated

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