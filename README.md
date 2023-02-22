# Autometrics Go

Autometrics generated automatically in Go.

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
// Somewhere in your file, probably at the top
//go:generate autometrics

//autometrics:doc
func RouteHandler(args interface{}) (err error) { // Name the error return value
        defer autometrics.Autometrics()
        // Do stuff
        return nil
}
```

It will automatically register the correct metrics to the default, global prometheus
metrics registry, which you can then expose as an extra route (like `/metrics`) to
your prometheus instance.

You will also get documentation generated with links to see metrics about your
function directly in your prometheus instance (default to `http://localhost:9090`)

## Non-design goals

### Handle functions that do not return error

Autometrics will report the success rate of a function only if the function
returns a named value `err` of type `error`. Some constraints might be relaxed
in further versions, but for now that's the alpha version goal.
