package main

import (
	"fmt"
	"log"
	"os"

	"github.com/autometrics-dev/autometrics-go/internal/doc"
)

func main() {
	fmt.Println(len(os.Args), os.Args)
	fileName := os.Getenv("GOFILE")
	promGenerator := doc.NewPrometheusDoc()
	if err := doc.TransformFile(fileName, promGenerator); err != nil {
		log.Fatalf("error transforming %s: %s", fileName, err)
	}
}
