# Web application instrumented in seconds

This is an example application to show how fast you can setup autometrics.

It shows the generator usage and sets up Prometheus to showcase the
"automatically generated links" feature.

## Quick start

``` sh
GOOS=linux GOARCH=amd64 go build -o web-server ./cmd/main.go
docker compose up -d
./poll_server
```

Then open [main](./cmd/main.go) in your editor and interact with the documentation links!

## Dependencies

In order to run this example you need:

- Go (at least 1.18)
- Docker
- Docker Compose

## Explanations

### Setup

The basic code used is in [main.go.bak](./cmd/main.go.bak) for demonstration purposes.
Note that the code has a `autometrics.Init()` method call that initialize the metrics, and
adds a `/metrics` handler to serve prometheus metrics

We then just used `go generate ./...` to generate the documentation strings and the
automatic metric collection calls (in defer statements)

### Building the docker image

Build the web-server for the image architecture:

```sh
GOOS=linux GOARCH=amd64 go build -o web-server ./cmd/main.go
docker compose build
```

### Start the services

In one terminal you can launch the stack and the small helper script to poll the the server:

```sh
docker compose up -d
./poll_server
```

### Check the links on Prometheus

The metrics won't appear immediately, due to Prometheus needing to poll them first, but after
approximatively 10s, you will see that the autometrics metrics get automatically filled by
the code. You just needed 2 lines of code and 1 comment per function to instrument everything.

### Original input

The "original" input file for the webserver (before the call to `go generate ./...`) can
be found [here](./cmd/main.go.bak)
