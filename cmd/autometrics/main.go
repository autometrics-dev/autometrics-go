package main

import (
	"log"
	"os"

	"github.com/autometrics-dev/autometrics-go/internal/doc"
)

func main() {
	fileName := os.Getenv("GOFILE")
	moduleName := os.Getenv("GOPACKAGE")
	promGenerator := doc.NewPrometheusDoc()

	if err := doc.TransformFile(fileName, moduleName, promGenerator); err != nil {
		log.Fatalf("error transforming %s: %s", fileName, err)
	}
}
