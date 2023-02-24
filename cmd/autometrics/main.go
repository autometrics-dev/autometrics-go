package main

import (
	"log"
	"os"

	"github.com/autometrics-dev/autometrics-go/internal/doc"
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

	if err := doc.TransformFile(fileName, moduleName, promGenerator); err != nil {
		log.Fatalf("error transforming %s: %s", fileName, err)
	}
}
