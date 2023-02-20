package main

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

	http.HandleFunc("/", errorable(indexHandler))
	http.HandleFunc("/random-error", errorable(randomErrorHandler))

	log.Println("binding on https://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

//go:generate go run github.com/autometrics-dev/autometrics-go/cmd/autometrics
func indexHandler(w http.ResponseWriter, _ *http.Request) error {
	if _, err := fmt.Fprintf(w, "Hello, World!\n"); err != nil {
		return err
	}

	return nil
}

var handlerError = errors.New("failed to handle request")

//go:generate go run github.com/autometrics-dev/autometrics-go/cmd/autometrics
func randomErrorHandler(w http.ResponseWriter, _ *http.Request) error {
	isErr := rand.Intn(2) == 0

	if isErr {
		return handlerError
	} else {
		w.WriteHeader(http.StatusOK)
		return nil
	}
}

func errorable(handler func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}
