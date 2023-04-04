package main

import (
	"log"
	"os"

	"github.com/autometrics-dev/autometrics-go/internal/doc"
	"github.com/autometrics-dev/autometrics-go/internal/generate"
	"github.com/autometrics-dev/autometrics-go/pkg/autometrics"
)

const prometheusAddressEnvironmentVariable = "AM_PROMETHEUS_URL"

func main() {
	fileName := os.Getenv("GOFILE")
	moduleName := os.Getenv("GOPACKAGE")

	prometheusUrl, envVarExists := os.LookupEnv(prometheusAddressEnvironmentVariable)
	if !envVarExists {
		prometheusUrl = doc.DefaultPrometheusInstanceUrl
	}
	promGenerator := doc.NewPrometheusDoc(prometheusUrl)

	implementation := autometrics.PROMETHEUS

	if err := generate.TransformFile(fileName, moduleName, promGenerator, implementation); err != nil {
		log.Fatalf("error transforming %s: %s", fileName, err)
	}
}
