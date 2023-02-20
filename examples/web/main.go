package web

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/random-error", randomErrorHandler)

	log.Fatal(http.ListenAndServe(":8080", nil))
}

//go:generate go run github.com/autometrics-dev/autometrics-go/cmd/autometrics -o main.go main.go
func indexHandler(w http.ResponseWriter, _ *http.Request) {
	if _, err := fmt.Fprintf(w, "Hello, World!\n"); err != nil {
		http.Error(w, "failed to write output", http.StatusInternalServerError)
	}
}

var handlerError = errors.New("failed to handle request")

//go:generate go run github.com/autometrics-dev/autometrics-go/cmd/autometrics -o main.go main.go
func randomErrorHandler(w http.ResponseWriter, _ *http.Request) {
	isErr := rand.Intn(1) == 0

	if isErr {
		http.Error(w, handlerError.Error(), http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}
