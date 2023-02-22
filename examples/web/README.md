# Web application instrumented in seconds

This is an example application to show how fast you can setup autometrics.

It shows the generator usage and sets up Prometheus to showcase the
"automatically generated links" feature.

## Quick start

``` sh
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

We then:

- used `go generate ./...` to generate the documentation strings
- added the `defer` snippets for the functions we want to instrument

### Hacks

Documenting all the non-standard hacks to get this showcase working.

#### go:generate cookie

In order to make the go:generate work with the current local copy of autometrics, there
have been a few changes to the cookie. It mentions a relative path to the executable:

```go
//go:generate ../autometrics
```

which assumes that a local copy of `autometrics` has been built from the root of the
repository:

``` sh
go build -o autometrics ./cmd/autometrics/main.go
cp autometrics examples/web
```

In practice, you should just have

``` go
//go:generate autometrics
```

after you did the `go get` command to obtain the version.

#### Update go mod for the local examples folder

- Set the url switcheroo in the _global_ gitconfig (the --local one doesn't work)
``` toml
[url "ssh://git@github.com/"]
        insteadOf = https://github.com/
```

- Set the env vars all the way to make `go mod tidy` happy
``` sh
GOPROXY=direct GOPRIVATE=github.com/autometrics-dev/autometrics-go go mod tidy
```

### Docker Compose

The `docker-compose.yml` file that comes with the example just sets up a
Prometheus instance that's reachable at `http://localhost:9090` from outside,
and a local image that runs the current [web server](./cmd/main.go) reachable at
`http://localhost:62086`.

The "original" input file for the webserver (before the call to `go generate ./...`) can
be found [here](./cmd/main.go.bak)
