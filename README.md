# Autometrics Go

Autometrics generated automatically in Go. This README is just being used as
a private, primary spec/exploration on what we want to do.

## Design Goals

Given a starting function like:

```go
func RouteHandler(args interface{}) error {
        // Do stuff
        return nil
}
```

Get metrics automatically generated changing your code to:
```go
//go:generate go-autometrics
func RouteHandler(args interface{}) (err error) { // Name the error return value
        defer RouteHandler_metrics()
        // Do stuff
        return nil
}
```

It will automatically register the correct metrics to the default, global prometheus
metrics registry, which you can then expose as an extra route (like `/metrics`) to
your prometheus instance

## Non-design goals

### Original source code modification

We don't want to modify your Go source code like `gomodifytags` does: you will
have to make the small patch above for each function that you want to instrument.

### Handle functions that do not return error

Autometrics will report the success rate of a function only if the function
returns a named value `err` of type `error`. Some constraints might be relaxed
in further versions, but for now that's the alpha version goal.

