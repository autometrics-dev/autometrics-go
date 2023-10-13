package main

import (
	"fmt"
	"log"
	"strings"

	internal "github.com/autometrics-dev/autometrics-go/internal/autometrics"
	"github.com/autometrics-dev/autometrics-go/internal/build"
	"github.com/autometrics-dev/autometrics-go/internal/generate"
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"

	arg "github.com/alexflint/go-arg"
)

const (
	DefaultPrometheusInstanceUrl = "http://localhost:9090/"
)

type args struct {
	FileName             string `arg:"-f,--,required,env:GOFILE" placeholder:"FILE_NAME" help:"File to transform."`
	ModuleName           string `arg:"-m,--,required,env:GOPACKAGE" placeholder:"MODULE_NAME" help:"Module containing the file to transform."`
	PrometheusUrl        string `arg:"--prom_url,env:AM_PROMETHEUS_URL" placeholder:"PROMETHEUS_URL" default:"http://localhost:9090" help:"Base URL of the Prometheus instance to generate links to."`
	UseOtel              bool   `arg:"--otel" default:"false" help:"Use OpenTelemetry client library to instrument code instead of default Prometheus."`
	AllowCustomLatencies bool   `arg:"--custom-latency" default:"false" help:"Allow non-default latencies to be used in latency-based SLOs."`
	DisableDocGeneration bool   `arg:"--no-doc,env:AM_NO_DOCGEN" default:"false" help:"Disable documentation links generation for all instrumented functions. Has the same effect as --no-doc in the //autometrics:inst directive."`
	ProcessAllFunctions  bool   `arg:"--inst-all,env:AM_INST_ALL" default:"false" help:"Instrument all function declared in the file to transform. Overwritten by the --rm-all argument if both are set."`
	RemoveAllFunctions   bool   `arg:"--rm-all,env:AM_RM_ALL" default:"false" help:"Remove all function instrumentation in the file to transform."`
}

func (args) Version() string {
	var buf strings.Builder

	fmt.Fprintf(&buf, "Autometrics %s", build.Version)

	return buf.String()
}

func (args) Description() string {
	var buf strings.Builder

	fmt.Fprintf(&buf,
		"Autometrics instruments annotated functions, and adds links in their doc comments to graphs of their live usage.\n\n")

	fmt.Fprintf(&buf,
		"It is meant to be used in a Go generator context. As such, it takes mandatory arguments in the form of environment variables.\n"+
			"You can also control the base URL of the prometheus instance in doc comments with an environment variable.\n")
	fmt.Fprintf(&buf,
		"\tNote: If you do not use the custom latencies in the SLO, the allowed latencies (in seconds) are %v\n\n",
		autometrics.DefBuckets)

	fmt.Fprintln(&buf,
		"Check https://github.com/autometrics-dev/autometrics-go for more help (including examples) and information.")
	fmt.Fprintf(&buf,
		"Autometrics is built by Fiberplane -- https://autometrics.dev\n")

	return buf.String()
}

func main() {
	var args args
	arg.MustParse(&args)

	implementation := autometrics.PROMETHEUS
	if args.UseOtel {
		implementation = autometrics.OTEL
	}

	ctx, err := internal.NewGeneratorContext(
		implementation,
		args.PrometheusUrl,
		args.AllowCustomLatencies,
		args.DisableDocGeneration,
		args.ProcessAllFunctions,
		args.RemoveAllFunctions,
	)
	if err != nil {
		log.Fatalf("error initialising autometrics context: %s", err)
	}

	if err := generate.TransformFile(ctx, args.FileName, args.ModuleName); err != nil {
		log.Fatalf("error transforming %s: %s", args.FileName, err)
	}
}
