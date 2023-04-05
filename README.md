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

### Import the libraries and initialize the metrics

In the main entrypoint of your program, you need to both add packages

``` go
import (
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	amImpl "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"
)
```

And then in your main function initialize the metrics

``` go
amImpl.Init(nil, autometrics.DefBuckets)
```
**_Warning_**: if you want to enable alerting from Autometrics, you **MUST NOT**
change the `autometrics.DefBuckets` argument. This is because alerting only uses
the bundled [Alerting rules file](./configs/autometrics.rules.yml).
Support for custom rule generation is planned but not present at the moment.

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
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
	amImpl "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)


func main() {
	amImpl.Init(nil, autometrics.DefBuckets)
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
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
-	amImpl "github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus"
+	amImpl "github.com/autometrics-dev/autometrics-go/pkg/autometrics/otel"
)
```
- change the call to `amImpl.Init` to the new signature: instead of a registry,
the `Init` function takes a meter name for the `otel_scope` label of the exported
metric. You can use the name of the application or its version for example

``` patch
-	amImpl.Init(nil, autometrics.DefBuckets)
+	amImpl.Init("myApp/v2/prod", autometrics.DefBuckets)
```

- add the `-otel` flag to the `//go:generate` directive

```patch
-//go:generate autometrics
+//go:generate autometrics -otel
```


### Implementation library import

## Status

The library is usable but not over, this section mentions the relevant points about
the current status

### Comments welcome

The first version of the library has _not_ been written by Go experts. Any comment or
code suggestion as Pull Request is more than welcome!

### Metrics system

For the time being only Prometheus metrics are supported, but the code has been
written with the possibility to have other systems, like OpenTelemetry,
integrated in the same way.
