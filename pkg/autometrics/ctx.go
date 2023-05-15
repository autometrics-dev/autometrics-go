package autometrics // import "github.com/autometrics-dev/autometrics-go/pkg/autometrics"

import (
	"context"
	"math/rand"
	"time"
)

type contextKey int

const (
	currentTraceId contextKey = iota
	currentSpanId
	parentSpanId
)

var randSource *rand.Rand

// Open Telemetry-compatible trace ID
type TraceID [16]byte

// Open Telemetry-compatible span ID
type SpanID [8]byte

// Context holds the configuration
// to instrument properly a function.
//
// This can be viewed as a context for the instrumentation calls
type Context struct {
	// Embedded context of the currently instrumented function.
	//
	// This allows the context to be passed around wherever a [context.Context] is expected.
	context.Context
	// TrackConcurrentCalls triggers the collection of the gauge for concurrent calls of the function.
	TrackConcurrentCalls bool
	// TrackCallerName adds a label with the caller name in all the collected metrics.
	TrackCallerName bool
	// AlertConf is an optional configuration to add alerting capabilities to the metrics.
	AlertConf *AlertConfiguration
	// StartTime is the start time of a single function execution.
	// Only amImpl.Instrument should read this value.
	// Only amImpl.PreInstrument should write this value.
	//
	// (amImpl is either the [Prometheus] or the [Open Telemetry] implementation)
	//
	// This value is only exported for the child packages [Prometheus] and [Open Telemetry]
	//
	// [Prometheus]: https://godoc.org/github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus
	// [Open Telemetry]: https://godoc.org/github.com/autometrics-dev/autometrics-go/pkg/autometrics/otel
	StartTime time.Time
	// CallInfo contains all the relevant data for caller information.
	// Only amImpl.Instrument should read this value.
	// Only amImpl.PreInstrument should write/read this value.
	//
	// (amImpl is either the [Prometheus] or the [Open Telemetry] implementation)
	//
	// This value is only exported for the child packages [Prometheus] and [Open Telemetry]
	//
	// [Prometheus]: https://godoc.org/github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus
	// [Open Telemetry]: https://godoc.org/github.com/autometrics-dev/autometrics-go/pkg/autometrics/otel
	CallInfo CallInfo
	// BuildInfo contains all the relevant data for caller information.
	// Only amImpl.Instrument and PreInstrument should read this value.
	// Only amImpl.Init should write/read this value.
	//
	// (amImpl is either the [Prometheus] or the [Open Telemetry] implementation)
	//
	// This value is only exported for the child packages [Prometheus] and [Open Telemetry]
	//
	// [Prometheus]: https://godoc.org/github.com/autometrics-dev/autometrics-go/pkg/autometrics/prometheus
	// [Open Telemetry]: https://godoc.org/github.com/autometrics-dev/autometrics-go/pkg/autometrics/otel
	BuildInfo BuildInfo
}

// NewContext is a constructor taking the parent context as argument.
//
// It accepts 'nil' as the parent context. In this case the constructor
// acts as if it received a new, fresh context.Background().
func NewContext(parentCtx context.Context) Context {
	if parentCtx == nil {
		parentCtx = context.Background()
	}
	return Context{
		TrackConcurrentCalls: true,
		TrackCallerName:      true,
		AlertConf:            nil,
		Context:              parentCtx,
	}
}

// SetTraceID sets the context's [TraceID]
func (ctx *Context) SetTraceID(tid TraceID) {
	ctx.Context = context.WithValue(ctx.Context, currentTraceId, tid)
}

// SetSpanID sets the context's [SpanID]
func (ctx *Context) SetSpanID(sid SpanID) {
	ctx.Context = context.WithValue(ctx.Context, currentSpanId, sid)
}

// SetParentSpanID sets the context's span's parent [SpanID]
func (ctx *Context) SetParentSpanID(sid SpanID) {
	ctx.Context = context.WithValue(ctx.Context, parentSpanId, sid)
}

// GetTraceID returns (_, false) if the context did not contain any trace id.
func (c Context) GetTraceID() (TraceID, bool) {
	if c.Context == nil {
		return TraceID{}, false
	}
	tid, ok := c.Value(currentTraceId).(TraceID)
	return tid, ok
}

// GetSpanID returns (_, false) if the context did not contain the current span id.
func (c Context) GetSpanID() (SpanID, bool) {
	if c.Context == nil {
		return SpanID{}, false
	}
	sid, ok := c.Value(currentSpanId).(SpanID)
	return sid, ok
}

// GetParentSpanID returns (_, false) if the context did not contain the parent's span id (including when we are in the root span).
func (c Context) GetParentSpanID() (SpanID, bool) {
	if c.Context == nil {
		return SpanID{}, false
	}
	sid, ok := c.Value(parentSpanId).(SpanID)
	return sid, ok
}

// FillTracingInfo ensures the context has a traceID and a spanID.
// If they do not have this information, this method adds randomly
// generated IDs in the context to be used later for exemplars
//
// The random generator is a PRNG, seeded with the timestamp of the first time new IDs are needed.
func (c *Context) FillTracingInfo() {
	// We are using a PRNG because FillTracingInfo is expected to be called in PreInstrument.
	// Therefore it can have a noticeable impact on the performance of instrumented code.
	// Pseudo randomness should be enough for our use cases, true randomness might introduce too much latency.
	// randSource is initialized with a timestamp from the first time it is accessed in nanoseconds, which should
	// be enough precision to avoid accidental collisions (imagine multiple services starting "at the same time" in a deployment).
	if randSource == nil {
		randSource = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	if parentSpanId, ok := c.GetSpanID(); ok {
		c.SetParentSpanID(parentSpanId)
	}

	sid := SpanID{}
	_, _ = randSource.Read(sid[:])
	c.SetSpanID(sid)

	// REVIEW: we might not want to fill the trace ID if it is absent.
	if _, ok := c.GetTraceID(); !ok {
		tid := TraceID{}
		_, _ = randSource.Read(tid[:])
		c.SetTraceID(tid)
	}
}

// GenerateTraceId generates a new TraceID with a Pseudo-random number generator.
//
// The generator is seeded with the timestamp of the first time a new traceID is needed.
func GenerateTraceId() TraceID {
	if randSource == nil {
		randSource = rand.New(rand.NewSource(time.Now().UnixNano()))
	}

	tid := TraceID{}
	randSource.Read(tid[:])

	return tid
}

// WithNewTraceId returns a copy of the passed context, with a newly generated traceID accessible for autometrics.
func WithNewTraceId(ctx context.Context) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, currentTraceId, GenerateTraceId())
}

// FillBuildInfo adds the relevant build information to the current context.
func (c *Context) FillBuildInfo() {
	c.BuildInfo.Version = GetVersion()
	c.BuildInfo.Commit = GetCommit()
	c.BuildInfo.Branch = GetBranch()
}
