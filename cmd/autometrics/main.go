package main

import (
	"fmt"
	"log"
	"os"

	internal "github.com/autometrics-dev/autometrics-go/internal/autometrics"
	"github.com/autometrics-dev/autometrics-go/internal/build"
	"github.com/autometrics-dev/autometrics-go/internal/generate"
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
)

const (
	prometheusAddressEnvironmentVariable = "AM_PROMETHEUS_URL"
	useOtelFlag                          = "-otel"
	allowCustomLatencies                 = "-custom-latency"
	shortVersionFlag                     = "-v"
	longVersionFlag                      = "-version"
	shortHelpFlag                        = "-h"
	longHelpFlag                         = "-help"
	DefaultPrometheusInstanceUrl         = "http://localhost:9090/"
)

func main() {
	fileName := os.Getenv("GOFILE")
	moduleName := os.Getenv("GOPACKAGE")
	args := os.Args

	if contains(args, longVersionFlag) || contains(args, shortVersionFlag) {
		printVersion()
		os.Exit(0)
	}

	if contains(args, longHelpFlag) || contains(args, shortHelpFlag) {
		printHelp()
		os.Exit(0)
	}

	if fileName == "" {
		printHelp()
		os.Exit(1)
	}

	prometheusUrl, envVarExists := os.LookupEnv(prometheusAddressEnvironmentVariable)
	if !envVarExists {
		prometheusUrl = DefaultPrometheusInstanceUrl
	}

	implementation := autometrics.PROMETHEUS
	if contains(args, useOtelFlag) {
		implementation = autometrics.OTEL
	}

	ctx, err := internal.NewGeneratorContext(implementation, prometheusUrl, contains(args, allowCustomLatencies))
	if err != nil {
		log.Fatalf("error initialising autometrics context: %s", err)
	}

	if err := generate.TransformFile(ctx, fileName, moduleName); err != nil {
		log.Fatalf("error transforming %s: %s", fileName, err)
	}
}

func contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}

func printVersion() {
	fmt.Printf("%s\n", build.Version)
	if build.Time != "" {
		fmt.Printf("Built on %s\n", build.Time)
	}
}

func printHelp() {
	fmt.Printf("Autometrics %s", build.Version)
	if build.Time != "" {
		fmt.Printf(" (%s)", build.Time)
	}
	fmt.Printf("\nBuilt by Autometrics team -- https://autometrics.dev\n\n")

	fmt.Printf(
		"usage: %s [%s | %s] [%s | %s] [%s] [%s] \n\n",
		os.Args[0],
		shortVersionFlag,
		longVersionFlag,
		shortHelpFlag,
		longHelpFlag,
		useOtelFlag,
		allowCustomLatencies,
	)

	fmt.Println("Autometrics is meant to be used in a Go generator context. As such, it takes mandatory arguments in the form of environment variables:")
	fmt.Println("    GOFILE\tPath to the file to transform.")
	fmt.Println("    GOPACKAGE\tName to the containing package.")
	fmt.Printf("\n\n")

	fmt.Println("Autometrics generates links to a prometheus instance in the doc comments of instrumented functions. You can control the base URL of the pointed to instance with an environment variable:")
	fmt.Printf("    %s\tBase URL of the Prometheus instance to generate links to (default: %s)\n",
		prometheusAddressEnvironmentVariable,
		DefaultPrometheusInstanceUrl,
	)
	fmt.Printf("\n\n")

	fmt.Printf("    %s\tUse OpenTelemetry client library to instrument code instead of default Prometheus\n",
		useOtelFlag,
	)
	fmt.Printf("    %s\tAllow non-default latencies to be used in latency-based SLOs (the default values in seconds are %v)\n",
		allowCustomLatencies,
		autometrics.DefBuckets,
	)
	fmt.Printf("\n")

	fmt.Println("Check https://github.com/autometrics-dev/autometrics-go for more help (including examples) and usage information.")
}
