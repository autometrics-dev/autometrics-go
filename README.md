# Autometrics Go

Autometrics is a [Go
Generator](https://pkg.go.dev/cmd/go#hdr-Generate_Go_files_by_processing_source)
bundled with a library that instruments your functions and gives direct links to 
inspect usage metrics from your code.

![Documentation comments of instrumented function is augmented with links](./assets/codium-screenshot-example.png)

You can optionnally add alerting rules so that code annotations make Prometheus
trigger alerts directly from production usage:

![a Slack bot is posting an alert directly in the channel](./assets/slack-alert-example.png)

A fully working use-case and example of library usage is available in the
[examples/web](./examples/web) subdirectory

## How to use

There is a one-time setup phase to prime the code for autometrics. Once this
phase is accomplished, only calling `go generate` is necessary.

### Install the go generator.

The generator is the binary in cmd/autometrics, so the easiest way to get it is
to install it through go:

```console
go install github.com/autometrics-dev/autometrics-go/cmd/autometrics@latest
```

### Import the libraries and initialize the metrics

In the main entrypoint of your program, you need to both add package

``` go
import (
	am "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"
)
```

And then in your main function initialize the metrics

``` go
am.Init(nil, am.DefBuckets)
```

> **Warning**
> If you want to enable alerting from Autometrics, you **MUST**
have the `--latency-ms` values to match the values given in your buckets. The
values in the buckets are given in _seconds_. By default, the generator will
error and tell you the valid default values if they don't match.
If the default values do not match you use case, you can change the buckets in
the init call, and add a `-custom-latency` argument to the `//go:generate` invocation.

```patch
-//go:generate autometrics
+//go:generate autometrics -custom-latency
```

### Add cookies in your code

Given a starting function like:

```go
func RouteHandler(args interface{}) error {
        // Do stuff
        return nil
}
```

The manual changes you need to do are:

```go
//go:generate autometrics

//autometrics:doc
func RouteHandler(args interface{}) (err error) { // Name the error return value; this is an optional but recommended change
        // Do stuff
        return nil
}
```

If you want the generated metrics to contain the function success rate, you
_must_ name the error return value. This is why we recommend to name the error
value you return for function you want to instrument.

### Generate the documentation and instrumentation code

Install the go generator using `go install` as usual:

``` console
go install https://github.com/autometrics-dev/autometrics-go/cmd/autometrics
```

Once you've done this, the `autometrics` generator takes care of the rest, and you can
simply call `go generate` with an optional environment variable:

```console
$ AM_PROMETHEUS_URL=http://localhost:9090/ go generate ./...
```

The generator will augment your doc comment to add quick links to metrics (using
the Prometheus URL as base URL), and add a unique defer statement that will take
care of instrumenting your code.

The environment variable `AM_PROMETHEUS_URL` controls the base URL of the instance that
is scraping the deployed version of your code. Having an environment variable means you
can change the generated links without touching your code. The default value, if absent,
is `http://localhost:9090/`.

You can have any value here, the only adverse impact it can
have is that the links in the doc comment might lead nowhere useful.

### Expose metrics outside

The last step now is to actually expose the generated metrics to the Prometheus instance.

For Prometheus the shortest way is to add the handler code in your main entrypoint:

``` go
import (
	am "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)


func main() {
	am.Init(nil, am.DefBuckets)
	http.Handle("/metrics", promhttp.Handler())
}
```

This is the shortest way to initialize and expose the metrics that autometrics will use
in the generated code.

### (OPTIONAL) Generate alerts automatically

Change the annotation of the function to automatically generate alerts for it:

``` go
//autometrics:doc --slo "Api" --success-target 90
func RouteHandler(args interface{}) (err error) {
        // Do stuff
        return nil
}
```

Then **you need to add** the [bundled](./configs/autometrics.rules.yml)
recording rules to your prometheus configuration.

The valid arguments for alert generation are:
- `--slo` (*MANDATORY*): name of the service for which the objective is relevant
- `--success-rate` : target success rate of the function, between 0 and 100 (you
  must name the `error` return value of the function for detection to work.)
- `--latency-ms` : maximum latency allowed for the function, in milliseconds.
- `--latency-target` : latency target for the threshold, between 0 and 100 (so X%
  of calls must last less than `latency-ms` milliseconds). You must specify both
  latency options, or none.
  
**_Warning_**: the generator will error out if you use targets that are not
supported by the bundled [Alerting rules file](./configs/autometrics.rules.yml).
Support for custom target is planned but not present at the moment
  
## (OPTIONAL) OpenTelemetry Support

Autometrics supports using OpenTelemetry with a prometheus exporter instead of using
Prometheus to publish the metrics. The changes you need to make are:

- change where the `amImpl` import points to
```patch
import (
-	am "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"
+	am "github.com/autometrics-dev/autometrics-go/pkg/autometrics/otel"
)
```
- change the call to `amImpl.Init` to the new signature: instead of a registry,
the `Init` function takes a meter name for the `otel_scope` label of the exported
metric. You can use the name of the application or its version for example

``` patch
-	am.Init(nil, am.DefBuckets)
+	am.Init("myApp/v2/prod", am.DefBuckets)
```

- add the `-otel` flag to the `//go:generate` directive

```patch
-//go:generate autometrics
+//go:generate autometrics -otel
```

## (OPTIONAL) Git hook

As autometrics is a Go generator that modifies the source code when run, it
might be interesting to set up `go generate ./...` to run in a git pre-commit
hook so that you never forget to run it if you change the source code.

If you use a tool like [pre-commit](https://pre-commit.com/), see their
documentation about how to add a hook that will run `go generate ./...`.

Otherwise, a simple example has been added in the [configs folder](./configs/pre-commit)
as an example. You can copy this file in your copy of your project's repository, within
`.git/hooks` and make sure that the file is executable.

## Status

The library is usable but not over, this section mentions the relevant points about
the current status

### Comments welcome

The first version of the library has _not_ been written by Go experts. Any comment or
code suggestion as Pull Request is more than welcome!

### Support for custom alerting rules generation

The alerting system for SLOs that Autometrics uses is based on
[Sloth](https://github.com/slok/sloth), and it has native Go types for
marshalling/unmarshalling rules, so it should be possible to provide an extra
binary in this repository, that only takes care of generating a new [rules
file](./configs/autometrics.rules.yml) with custom objectives.
