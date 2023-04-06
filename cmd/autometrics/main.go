package main

import (
	"log"
	"os"

	internal "github.com/autometrics-dev/autometrics-go/internal/autometrics"
	"github.com/autometrics-dev/autometrics-go/internal/generate"
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
)

const (
	prometheusAddressEnvironmentVariable = "AM_PROMETHEUS_URL"
	useOtelFlag                          = "-otel"
	allowCustomLatencies                 = "-custom-latency"
	DefaultPrometheusInstanceUrl         = "http://localhost:9090/"
)

func main() {
	fileName := os.Getenv("GOFILE")
	moduleName := os.Getenv("GOPACKAGE")
	args := os.Args

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
